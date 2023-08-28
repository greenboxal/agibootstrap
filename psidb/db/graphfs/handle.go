package graphfs

import (
	"context"
	"io"
	"io/fs"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type NodeHandleOperations interface {
	Read(ctx context.Context, nh NodeHandle) (*SerializedNode, error)
	Write(ctx context.Context, nh NodeHandle, fe *SerializedNode) error

	SetEdge(ctx context.Context, nh NodeHandle, edge *SerializedEdge) error
	RemoveEdge(ctx context.Context, nh NodeHandle, key psi.EdgeKey) error

	ReadEdge(ctx context.Context, nh NodeHandle, key psi.EdgeKey) (*SerializedEdge, error)
	ReadEdges(ctx context.Context, nh NodeHandle) (iterators.Iterator[*SerializedEdge], error)
}

type NodeHandle interface {
	Transaction() *Transaction
	Inode() *INode
	Entry() *CacheEntry
	Options() OpenNodeOptions

	Read(ctx context.Context) (*SerializedNode, error)
	Write(ctx context.Context, fe *SerializedNode) error

	SetEdge(ctx context.Context, edge *SerializedEdge) error
	RemoveEdge(ctx context.Context, key psi.EdgeKey) error

	ReadEdge(ctx context.Context, key psi.EdgeKey) (*SerializedEdge, error)
	ReadEdges(ctx context.Context) (iterators.Iterator[*SerializedEdge], error)

	io.Closer
}

var ErrHandleClosed = errors.New("handle closed")

type NodeHandleBase struct {
	inode   *INode
	dentry  *CacheEntry
	closed  bool
	options OpenNodeOptions
}

func (nh *NodeHandleBase) Transaction() *Transaction { return nh.options.Transaction }
func (nh *NodeHandleBase) Inode() *INode             { return nh.inode }
func (nh *NodeHandleBase) Entry() *CacheEntry        { return nh.dentry }
func (nh *NodeHandleBase) Options() OpenNodeOptions  { return nh.options }

func (nh *NodeHandleBase) Read(ctx context.Context) (*SerializedNode, error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	if nh.options.Flags&OpenNodeFlagsRead == 0 {
		return nil, fs.ErrPermission
	}

	nh.inode.mu.RLock()
	defer nh.inode.mu.RUnlock()

	nh.inode.lastVersionMutex.RLock()
	sn := nh.inode.lastVersion
	nh.inode.lastVersionMutex.RUnlock()

	if sn != nil && sn.Flags&NodeFlagInvalid == 0 {
		if sn.Flags&NodeFlagRemoved != 0 {
			return nil, psi.ErrNodeNotFound
		}

		return sn, nil
	}

	nh.inode.lastVersionMutex.Lock()
	defer nh.inode.lastVersionMutex.Unlock()

	sn, err := nh.inode.NodeHandleOperations().Read(ctx, nh)

	if err != nil {
		return nil, err
	}

	if sn.Flags&NodeFlagRemoved != 0 {
		return nil, psi.ErrNodeNotFound
	}

	nh.inode.lastVersion = sn

	return sn, nil
}

func (nh *NodeHandleBase) Write(ctx context.Context, fe *SerializedNode) error {
	if nh.closed {
		return ErrHandleClosed
	}

	if nh.options.Flags&OpenNodeFlagsWrite == 0 {
		return fs.ErrPermission
	}

	nh.inode.mu.RLock()
	defer nh.inode.mu.RUnlock()

	nh.inode.lastVersionMutex.Lock()
	defer nh.inode.lastVersionMutex.Unlock()

	if err := nh.inode.NodeHandleOperations().Write(ctx, nh, fe); err != nil {
		return err
	}

	frozen := *fe
	frozen.Flags &= ^NodeFlagInvalid
	frozen.Flags &= ^NodeFlagRemoved
	nh.inode.lastVersion = &frozen

	return nil
}

func (nh *NodeHandleBase) SetEdge(ctx context.Context, edge *SerializedEdge) error {
	if nh.closed {
		return ErrHandleClosed
	}

	if nh.options.Flags&OpenNodeFlagsWrite == 0 {
		return fs.ErrPermission
	}

	return nh.inode.NodeHandleOperations().SetEdge(ctx, nh, edge)
}

func (nh *NodeHandleBase) RemoveEdge(ctx context.Context, key psi.EdgeKey) error {
	if nh.closed {
		return ErrHandleClosed
	}

	if nh.options.Flags&OpenNodeFlagsWrite == 0 {
		return fs.ErrPermission
	}

	return nh.inode.NodeHandleOperations().RemoveEdge(ctx, nh, key)
}

func (nh *NodeHandleBase) ReadEdge(ctx context.Context, key psi.EdgeKey) (*SerializedEdge, error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	if nh.options.Flags&OpenNodeFlagsRead == 0 {
		return nil, fs.ErrPermission
	}

	return nh.inode.NodeHandleOperations().ReadEdge(ctx, nh, key)
}

func (nh *NodeHandleBase) ReadEdges(ctx context.Context) (iterators.Iterator[*SerializedEdge], error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	if nh.options.Flags&OpenNodeFlagsRead == 0 {
		return nil, fs.ErrPermission
	}

	return nh.inode.NodeHandleOperations().ReadEdges(ctx, nh)
}

func (nh *NodeHandleBase) Close() error {
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

func NewNodeHandle(ctx context.Context, inode *INode, dentry *CacheEntry, options OpenNodeOptions) (NodeHandle, error) {
	base := &NodeHandleBase{
		inode:   inode.Get(),
		dentry:  dentry.Get(),
		options: options,
	}

	if options.Transaction != nil {
		return &txNodeHandle{
			tx:      options.Transaction,
			inode:   inode.Get(),
			dentry:  dentry.Get(),
			options: options,

			baseHandle: base,
		}, nil
	}

	return base, nil
}
