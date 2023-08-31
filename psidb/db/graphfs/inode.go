package graphfs

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/greenboxal/agibootstrap/psidb/core/api"
)

type INodeOperations interface {
	Create(ctx context.Context, self *CacheEntry, options OpenNodeOptions) error

	Lookup(ctx context.Context, self *INode, dentry *CacheEntry) (*CacheEntry, error)

	Unlink(ctx context.Context, self *INode, dentry *CacheEntry) error
}

type INodeFlag uint64

const (
	INodeFlagReferenced INodeFlag = 1 << iota
	INodeFlagNew

	INodeFlagNone INodeFlag = 0
)

type INode struct {
	mu sync.RWMutex

	id           int64
	refcount     atomic.Int32
	flags        INodeFlag
	cacheEntries []*CacheEntry

	lastVersionMutex sync.RWMutex
	lastVersion      *coreapi.SerializedNode
	edgeCache        map[string]*coreapi.SerializedEdge

	sb SuperBlock
}

// AllocateInode allocates a new inode with the given id, iops, and nops.
func AllocateInode(sb SuperBlock, id int64) *INode {
	return &INode{
		sb: sb,
		id: id,
	}
}

func (i *INode) ID() int64                        { return i.id }
func (i *INode) SuperBlock() SuperBlock           { return i.sb }
func (i *INode) INodeOperations() INodeOperations { return i.sb.INodeOperations() }
func (i *INode) NodeHandleOperations() NodeHandleOperations {
	return i.sb.NodeHandleOperations()
}

func (i *INode) IsValid() bool { return i.flags&INodeFlagReferenced != 0 }

func (i *INode) SetFlags(f INodeFlag) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.flags |= f
}

func (i *INode) ClearFlags(f INodeFlag) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.flags &= ^f
}

func (i *INode) Flags() INodeFlag {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.flags
}

func (i *INode) attachCacheEntry(ce *CacheEntry) {
	i.mu.Lock()
	defer i.mu.Unlock()

	candidateIdx := -1

	for i, e := range i.cacheEntries {
		if e == ce {
			return
		} else if e == nil && candidateIdx == -1 {
			candidateIdx = i
		}
	}

	if candidateIdx != -1 {
		i.cacheEntries[candidateIdx] = ce
	} else {
		i.cacheEntries = append(ce.child, ce)
	}
}

func (i *INode) detachCacheEntry(ce *CacheEntry) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for idx, e := range i.cacheEntries {
		if e == ce {
			i.cacheEntries[idx] = nil

			return
		}
	}
}

func (i *INode) delete() {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.refcount.Load() > 0 {
		return
	}
}

func (i *INode) Get() *INode {
	i.Ref()

	return i
}

func (i *INode) Ref() {
	i.refcount.Add(1)
	i.flags |= INodeFlagReferenced
}

func (i *INode) Unref() {
	if i.refcount.Add(-1) == 0 {
		i.delete()
	}
}
