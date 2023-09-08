package graphfs

import (
	"context"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type txNode struct {
	Inode  int64
	Frozen coreapi.SerializedNode
	Edges  map[string]coreapi.SerializedEdge
}

type txNodeHandle struct {
	tx      *Transaction
	inode   *INode
	dentry  *CacheEntry
	closed  bool
	options OpenNodeOptions

	baseHandle NodeHandle
}

func (nh *txNodeHandle) Transaction() *Transaction { return nh.tx }
func (nh *txNodeHandle) Inode() *INode             { return nh.inode }
func (nh *txNodeHandle) Entry() *CacheEntry        { return nh.dentry }
func (nh *txNodeHandle) Options() OpenNodeOptions  { return nh.options }

func (nh *txNodeHandle) getOrCreateStagedNode() *txNode {
	n := nh.tx.dirtyNodes[nh.inode.id]

	if n == nil {
		n = &txNode{
			Inode: nh.inode.id,
			Frozen: coreapi.SerializedNode{
				Flags: coreapi.NodeFlagInvalid,
			},
			Edges: map[string]coreapi.SerializedEdge{},
		}

		nh.tx.dirtyNodes[nh.inode.id] = n
	}

	return n
}

func (nh *txNodeHandle) Read(ctx context.Context) (*coreapi.SerializedNode, error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	staged := nh.getOrCreateStagedNode()

	if staged.Frozen.Flags&coreapi.NodeFlagRemoved != 0 {
		return nil, psi.ErrNodeNotFound
	}

	if staged.Frozen.Flags&coreapi.NodeFlagInvalid == 0 {
		return &staged.Frozen, nil
	}

	sn, err := nh.inode.NodeHandleOperations().Read(ctx, nh)

	if err != nil {
		return nil, err
	}

	staged.Frozen = *sn

	return sn, nil
}

func (nh *txNodeHandle) Write(ctx context.Context, fe *coreapi.SerializedNode) error {
	if nh.closed {
		return ErrHandleClosed
	}

	if err := nh.tx.Append(ctx, coreapi.JournalEntry{
		Op:    coreapi.JournalOpWrite,
		Inode: nh.inode.id,
		Path:  &fe.Path,
		Node:  fe,
	}); err != nil {
		return err
	}

	staged := nh.getOrCreateStagedNode()
	staged.Frozen = *fe
	staged.Frozen.Flags &= ^coreapi.NodeFlagInvalid
	staged.Frozen.Flags &= ^coreapi.NodeFlagRemoved

	return nil
}

func (nh *txNodeHandle) SetEdge(ctx context.Context, edge *coreapi.SerializedEdge) error {
	if nh.closed {
		return ErrHandleClosed
	}

	p := nh.dentry.Path()

	if err := nh.tx.Append(ctx, coreapi.JournalEntry{
		Op:    coreapi.JournalOpSetEdge,
		Inode: nh.inode.id,
		Path:  &p,
		Edge:  edge,
	}); err != nil {
		return err
	}

	e := *edge
	e.Flags &= ^coreapi.EdgeFlagRemoved

	staged := nh.getOrCreateStagedNode()
	staged.Edges[edge.Key.String()] = e

	return nil
}

func (nh *txNodeHandle) RemoveEdge(ctx context.Context, key psi.EdgeKey) error {
	if nh.closed {
		return ErrHandleClosed
	}

	p := nh.dentry.Path()

	staged := nh.getOrCreateStagedNode()
	se, ok := staged.Edges[key.String()]

	if !ok {
		se = coreapi.SerializedEdge{Key: key, Flags: coreapi.EdgeFlagRemoved}
	}

	se.Flags |= coreapi.EdgeFlagRemoved

	if err := nh.tx.Append(ctx, coreapi.JournalEntry{
		Op:    coreapi.JournalOpRemoveEdge,
		Inode: nh.inode.id,
		Path:  &p,
		Edge:  &se,
	}); err != nil {
		return err
	}

	staged.Edges[se.Key.String()] = se

	return nil
}

func (nh *txNodeHandle) ReadEdge(ctx context.Context, key psi.EdgeKey) (*coreapi.SerializedEdge, error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	n := nh.getOrCreateStagedNode()

	if e, ok := n.Edges[key.String()]; ok {
		if e.Flags&coreapi.EdgeFlagRemoved != 0 {
			return nil, psi.ErrNodeNotFound
		}

		return &e, nil
	}

	e, err := nh.inode.NodeHandleOperations().ReadEdge(ctx, nh, key)

	if err != nil {
		return nil, err
	}

	if e.Flags&coreapi.EdgeFlagRemoved != 0 {
		return nil, psi.ErrNodeNotFound
	}

	n.Edges[key.String()] = *e

	return e, nil
}

func (nh *txNodeHandle) ReadEdges(ctx context.Context) (iterators.Iterator[*coreapi.SerializedEdge], error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	n := nh.getOrCreateStagedNode()

	base, err := nh.inode.NodeHandleOperations().ReadEdges(ctx, nh)

	if err != nil && (errors.Is(err, psi.ErrNodeNotFound) && n == nil) {
		return nil, err
	}

	if base == nil {
		base = iterators.Empty[*coreapi.SerializedEdge]()
	}

	if n == nil {
		return base, nil
	}

	dirtyEdges := iterators.FromMap(n.Edges)
	seenMap := map[string]struct{}{}

	return iterators.NewIterator(func() (*coreapi.SerializedEdge, bool) {
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

				if e.Flags&coreapi.EdgeFlagRemoved != 0 {
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

				if e.V.Flags&coreapi.EdgeFlagRemoved != 0 {
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
