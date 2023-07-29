package graphfs

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

// CacheEntry is similar to linux VFS in the context of dcache/dentry.

type CacheEntryFlag uint64

const (
	// CacheEntryFlagReferenced is set when the cache entry is referenced.
	CacheEntryFlagReferenced CacheEntryFlag = 1 << iota
	// CacheEntryFlagNegative is set when the cache entry is negative.
	CacheEntryFlagNegative
	// CacheEntryFlagDisconnected is set when the cache entry is disconnected.
	CacheEntryFlagDisconnected

	CacheEntryFlagNone CacheEntryFlag = 0
)

// CacheEntry is similar to linux VFS in the context of dcache/dentry.
type CacheEntry struct {
	mu sync.RWMutex

	flags    CacheEntryFlag
	refcount atomic.Int32

	sb    SuperBlock
	inode *INode

	parent *CacheEntry
	name   psi.PathElement
	child  []*CacheEntry
}

// AllocCacheEntryRoot allocates a root cache entry.
func AllocCacheEntryRoot(sb SuperBlock) *CacheEntry {
	return AllocCacheEntry(sb, nil, psi.PathElement{})
}

// AllocCacheEntry allocates a cache entry.
func AllocCacheEntry(sb SuperBlock, parent *CacheEntry, name psi.PathElement) *CacheEntry {
	if parent != nil {
		parent = parent.Get()
	}

	return &CacheEntry{
		sb:     sb,
		name:   name,
		parent: parent,
		flags:  CacheEntryFlagDisconnected | CacheEntryFlagNegative,
	}
}

// Name returns the name of the cache entry.
func (ce *CacheEntry) Name() psi.PathElement { return ce.name }

// Path returns the path of the cache entry.
func (ce *CacheEntry) Path() psi.Path {
	if ce.parent == nil {
		return psi.PathFromElements(ce.sb.UUID(), false, ce.name)
	}

	return ce.parent.Path().Child(ce.name)
}

// Parent returns the parent cache entry.
func (ce *CacheEntry) Parent() *CacheEntry { return ce.parent }

// Inode returns the inode of the cache entry.
func (ce *CacheEntry) Inode() *INode { return ce.inode }

// IsValid returns true if the cache entry is valid.
func (ce *CacheEntry) IsValid() bool { return ce.flags&CacheEntryFlagReferenced != 0 }

// IsNegative returns true if the cache entry is negative.
func (ce *CacheEntry) IsNegative() bool { return ce.flags&CacheEntryFlagNegative != 0 }

// IsDisconnected returns true if the cache entry is disconnected.
func (ce *CacheEntry) IsDisconnected() bool { return ce.flags&CacheEntryFlagDisconnected != 0 }

func (ce *CacheEntry) INodeOperations() INodeOperations { return ce.sb.INodeOperations() }

// Instantiate instantiates the cache entry. The cache entry must be
// disconnected.
func (ce *CacheEntry) Instantiate(inode *INode) {
	if ce.flags&CacheEntryFlagDisconnected != 0 {
		panic("CacheEntry.Instantiate: cache entry is disconnected")
	}

	if ce.inode == inode {
		return
	}

	ce.inode = inode.Get()

	if ce.inode != nil {
		ce.flags &= ^CacheEntryFlagNegative

		inode.attachCacheEntry(ce)
	} else {
		ce.flags |= CacheEntryFlagNegative
	}
}

// Add adds the inode to the cache entry. The cache entry must be disconnected.
// The inode must be valid.
func (ce *CacheEntry) Add(inode *INode) {
	if ce.flags&CacheEntryFlagDisconnected == 0 {
		panic("CacheEntry.Add: cache entry is not disconnected")
	}

	ce.mu.Lock()
	defer ce.mu.Unlock()

	if ce.parent != nil && ce.parent != ce {
		ce.parent.addChild(ce)
	}

	ce.flags &= ^CacheEntryFlagDisconnected

	ce.Instantiate(inode)
}

// Lookup looks up the cache entry.
func (ce *CacheEntry) Lookup(ctx context.Context, name psi.PathElement) (*CacheEntry, error) {
	for _, c := range ce.child {
		if c == nil {
			continue
		}

		if c.name == name {
			return c, nil
		}
	}

	child := AllocCacheEntry(ce.sb, ce, name)

	return ce.sb.INodeOperations().Lookup(ctx, ce.inode, child)
}

// delete deletes the cache entry after it is unreferenced.
func (ce *CacheEntry) delete() {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	if ce.refcount.Load() > 0 {
		return
	}

	if ce.parent != nil {
		ce.parent.deleteChild(ce)
		ce.parent.Unref()
		ce.parent = nil

		ce.flags |= CacheEntryFlagDisconnected
	}

	if ce.inode != nil {
		ce.inode.detachCacheEntry(ce)
		ce.inode.Unref()
		ce.inode = nil

		ce.flags |= CacheEntryFlagNegative
	}

	ce.flags &= ^CacheEntryFlagReferenced
}

// addChild adds the child to the cache entry.
func (ce *CacheEntry) addChild(child *CacheEntry) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	candidateIdx := -1

	for i, c := range ce.child {
		if c == nil && candidateIdx == -1 {
			candidateIdx = i
		} else if c == child {
			return
		}
	}

	if candidateIdx != -1 {
		ce.child[candidateIdx] = child
	} else {
		ce.child = append(ce.child, child)
	}
}

func (ce *CacheEntry) deleteChild(child *CacheEntry) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	for i, c := range ce.child {
		if c == child {
			ce.child[i] = nil
			return
		}
	}
}

func (ce *CacheEntry) Get() *CacheEntry {
	ce.ref()

	return ce
}

func (ce *CacheEntry) ref() {
	ce.refcount.Add(1)
}

func (ce *CacheEntry) Unref() {
	if ce.refcount.Add(-1) == 0 {
		ce.delete()
	}
}

func (ce *CacheEntry) MakeNegative() {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	ce.flags |= CacheEntryFlagNegative
}
