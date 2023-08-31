package graphfs

import (
	"context"
	"io"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
)

type NodeHandleOperations interface {
	Read(ctx context.Context, nh NodeHandle) (*coreapi.SerializedNode, error)
	Write(ctx context.Context, nh NodeHandle, fe *coreapi.SerializedNode) error

	SetEdge(ctx context.Context, nh NodeHandle, edge *coreapi.SerializedEdge) error
	RemoveEdge(ctx context.Context, nh NodeHandle, key psi.EdgeKey) error

	ReadEdge(ctx context.Context, nh NodeHandle, key psi.EdgeKey) (*coreapi.SerializedEdge, error)
	ReadEdges(ctx context.Context, nh NodeHandle) (iterators.Iterator[*coreapi.SerializedEdge], error)
}

type NodeHandle interface {
	Transaction() *Transaction
	Inode() *INode
	Entry() *CacheEntry
	Options() OpenNodeOptions

	Read(ctx context.Context) (*coreapi.SerializedNode, error)
	Write(ctx context.Context, fe *coreapi.SerializedNode) error

	SetEdge(ctx context.Context, edge *coreapi.SerializedEdge) error
	RemoveEdge(ctx context.Context, key psi.EdgeKey) error

	ReadEdge(ctx context.Context, key psi.EdgeKey) (*coreapi.SerializedEdge, error)
	ReadEdges(ctx context.Context) (iterators.Iterator[*coreapi.SerializedEdge], error)

	io.Closer
}

var ErrHandleClosed = errors.New("handle closed")
