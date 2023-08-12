package graphfs

import (
	"context"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type txNode struct {
	Inode  int64
	Frozen SerializedNode
	Edges  map[string]SerializedEdge
}

type txNodeHandle struct {
	tx      *Transaction
	inode   *INode
	dentry  *CacheEntry
	closed  bool
	options OpenNodeOptions
}

func (nh *txNodeHandle) Transaction() *Transaction { return nh.tx }
func (nh *txNodeHandle) Inode() *INode             { return nh.inode }
func (nh *txNodeHandle) Entry() *CacheEntry        { return nh.dentry }
func (nh *txNodeHandle) Options() OpenNodeOptions  { return nh.options }

func (nh *txNodeHandle) Read(ctx context.Context) (*SerializedNode, error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	n := nh.tx.dirtyNodes[nh.inode.id]

	if n != nil {
		if n.Frozen.Flags&NodeFlagRemoved != 0 {
			return nil, psi.ErrNodeNotFound
		}
	}

	return nh.inode.NodeHandleOperations().Read(ctx, nh)
}

func (nh *txNodeHandle) Write(ctx context.Context, fe *SerializedNode) error {
	if nh.closed {
		return ErrHandleClosed
	}

	return nh.tx.Append(ctx, JournalEntry{
		Op:    JournalOpWrite,
		Inode: nh.inode.id,
		Path:  &fe.Path,
		Node:  fe,
	})
}

func (nh *txNodeHandle) SetEdge(ctx context.Context, edge *SerializedEdge) error {
	if nh.closed {
		return ErrHandleClosed
	}

	p := nh.dentry.Path()

	return nh.tx.Append(ctx, JournalEntry{
		Op:    JournalOpSetEdge,
		Inode: nh.inode.id,
		Path:  &p,
		Edge:  edge,
	})
}

func (nh *txNodeHandle) RemoveEdge(ctx context.Context, key psi.EdgeKey) error {
	if nh.closed {
		return ErrHandleClosed
	}

	p := nh.dentry.Path()

	return nh.tx.Append(ctx, JournalEntry{
		Op:    JournalOpRemoveEdge,
		Inode: nh.inode.id,
		Path:  &p,
		Edge:  &SerializedEdge{Key: key},
	})
}

func (nh *txNodeHandle) ReadEdge(ctx context.Context, key psi.EdgeKey) (*SerializedEdge, error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	n := nh.tx.dirtyNodes[nh.inode.id]

	if n != nil {
		e, ok := n.Edges[key.String()]

		if ok {
			if e.Flags&EdgeFlagRemoved != 0 {
				return nil, psi.ErrNodeNotFound
			}

			return &e, nil
		}
	}

	return nh.inode.NodeHandleOperations().ReadEdge(ctx, nh, key)
}

func (nh *txNodeHandle) ReadEdges(ctx context.Context) (iterators.Iterator[*SerializedEdge], error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	n := nh.tx.dirtyNodes[nh.inode.id]

	base, err := nh.inode.NodeHandleOperations().ReadEdges(ctx, nh)

	if err != nil && (errors.Is(err, psi.ErrNodeNotFound) && n == nil) {
		return nil, err
	}

	if base == nil {
		base = iterators.Empty[*SerializedEdge]()
	}

	if n == nil {
		return base, nil
	}

	dirtyEdges := iterators.FromMap(n.Edges)
	seenMap := map[string]struct{}{}

	return iterators.NewIterator(func() (*SerializedEdge, bool) {
		for {
			if base != nil {
				if !base.Next() {
					base = nil
					continue
				}

				e := base.Value()

				if e2, ok := n.Edges[e.Key.String()]; ok {
					e = &e2
					seenMap[e.Key.String()] = struct{}{}
				}

				if e.Flags&EdgeFlagRemoved != 0 {
					continue
				}

				return e, true
			} else {
				if !dirtyEdges.Next() {
					return nil, false
				}

				e := dirtyEdges.Value()

				if _, ok := seenMap[e.K]; ok {
					continue
				}

				if e.V.Flags&EdgeFlagRemoved != 0 {
					continue
				}

				return &e.V, true
			}
		}
	}), nil
}

func (nh *txNodeHandle) Close() error {
	if nh.closed {
		return nil
	}

	if nh.dentry != nil {
		nh.dentry.Unref()
		nh.dentry = nil
	}

	if nh.inode != nil {
		nh.inode.Unref()
		nh.inode = nil
	}

	nh.closed = true

	return nil
}
