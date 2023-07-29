package graphfs

import (
	"context"
	"io"

	"github.com/ipld/go-ipld-prime"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type NodeHandleOperations interface {
	Read(ctx context.Context, nh NodeHandle) (*SerializedNode, error)
	Write(ctx context.Context, nh NodeHandle, fe *SerializedNode) (ipld.Link, error)

	SetEdge(ctx context.Context, nh NodeHandle, edge *SerializedEdge) error
	RemoveEdge(ctx context.Context, nh NodeHandle, key psi.EdgeKey) error

	ReadEdge(ctx context.Context, nh NodeHandle, key psi.EdgeKey) (*SerializedEdge, error)
	ReadEdges(ctx context.Context, nh NodeHandle) (iterators.Iterator[*SerializedEdge], error)
}

type NodeHandle interface {
	Inode() *INode
	Entry() *CacheEntry
	Options() OpenNodeOptions

	Read(ctx context.Context) (*SerializedNode, error)
	Write(ctx context.Context, fe *SerializedNode) (ipld.Link, error)

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

func (nh *NodeHandleBase) Read(ctx context.Context) (*SerializedNode, error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	return nh.inode.NodeHandleOperations().Read(ctx, nh)
}

func (nh *NodeHandleBase) Write(ctx context.Context, fe *SerializedNode) (ipld.Link, error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	return nh.inode.NodeHandleOperations().Write(ctx, nh, fe)
}

func (nh *NodeHandleBase) SetEdge(ctx context.Context, edge *SerializedEdge) error {
	if nh.closed {
		return ErrHandleClosed
	}

	return nh.inode.NodeHandleOperations().SetEdge(ctx, nh, edge)
}

func (nh *NodeHandleBase) RemoveEdge(ctx context.Context, key psi.EdgeKey) error {
	if nh.closed {
		return ErrHandleClosed
	}

	return nh.inode.NodeHandleOperations().RemoveEdge(ctx, nh, key)
}

func (nh *NodeHandleBase) ReadEdge(ctx context.Context, key psi.EdgeKey) (*SerializedEdge, error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	return nh.inode.NodeHandleOperations().ReadEdge(ctx, nh, key)
}

func (nh *NodeHandleBase) ReadEdges(ctx context.Context) (iterators.Iterator[*SerializedEdge], error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	return nh.inode.NodeHandleOperations().ReadEdges(ctx, nh)
}

func (nh *NodeHandleBase) Inode() *INode            { return nh.inode }
func (nh *NodeHandleBase) Entry() *CacheEntry       { return nh.dentry }
func (nh *NodeHandleBase) Options() OpenNodeOptions { return nh.options }

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

func NewNodeHandle(inode *INode, dentry *CacheEntry, options OpenNodeOptions) NodeHandle {
	return &NodeHandleBase{
		inode:   inode.Get(),
		dentry:  dentry.Get(),
		closed:  false,
		options: options,
	}
}
