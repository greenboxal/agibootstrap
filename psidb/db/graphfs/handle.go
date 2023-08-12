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

	return nh.inode.NodeHandleOperations().Read(ctx, nh)
}

func (nh *NodeHandleBase) Write(ctx context.Context, fe *SerializedNode) error {
	if nh.closed {
		return ErrHandleClosed
	}

	if nh.options.Flags&OpenNodeFlagsWrite == 0 {
		return fs.ErrPermission
	}

	return nh.inode.NodeHandleOperations().Write(ctx, nh, fe)
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
	if options.Transaction != nil {
		return &txNodeHandle{
			tx:      options.Transaction,
			inode:   inode.Get(),
			dentry:  dentry.Get(),
			options: options,
		}, nil
	}

	return &NodeHandleBase{
		inode:   inode.Get(),
		dentry:  dentry.Get(),
		options: options,
	}, nil
}
