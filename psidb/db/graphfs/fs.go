package graphfs

import (
	"context"

	"github.com/ipld/go-ipld-prime"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type FileSystemOperations interface {
	Open(ctx context.Context, path psi.Path, options ...OpenNodeOption) (NodeHandle, error)
	Resolve(ctx context.Context, path psi.Path) (*CacheEntry, error)

	Read(ctx context.Context, path psi.Path) (*psi.FrozenNode, error)
	Write(ctx context.Context, path psi.Path, node *psi.FrozenNode) (ipld.Link, error)

	ReadEdges(ctx context.Context, path psi.Path) (iterators.Iterator[*psi.FrozenEdge], error)
}

type OpenNodeFlags uint32

const (
	OpenNodeFlagsNone OpenNodeFlags = 0
	OpenNodeFlagsRead OpenNodeFlags = 1 << iota
	OpenNodeFlagsWrite
	OpenNodeFlagsCreate
	OpenNodeFlagsAppend
	OpenNodeFlagsTruncate
	OpenNodeFlagsExclusive
)

type OpenNodeOptions struct {
	Flags      OpenNodeFlags
	ForceInode *int64
}

func (o *OpenNodeOptions) Apply(opts ...OpenNodeOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func NewOpenNodeOptions(opts ...OpenNodeOption) OpenNodeOptions {
	o := OpenNodeOptions{}
	o.Apply(opts...)
	return o
}

type OpenNodeOption func(*OpenNodeOptions)

func WithOpenNodeCreateIfMissing() OpenNodeOption {
	return WithOpenNodeFlags(OpenNodeFlagsWrite | OpenNodeFlagsRead | OpenNodeFlagsCreate | OpenNodeFlagsAppend)
}

func WithOpenNodeFlag(flag OpenNodeFlags) OpenNodeOption {
	return func(opts *OpenNodeOptions) {
		opts.Flags |= flag
	}
}

func WithOpenNodeOptions(opts OpenNodeOptions) OpenNodeOption {
	return func(o *OpenNodeOptions) {
		*o = opts
	}
}

func WithOpenNodeFlags(flags OpenNodeFlags) OpenNodeOption {
	return func(opts *OpenNodeOptions) {
		opts.Flags = flags
	}
}

func WithOpenNodeForceInode(inode int64) OpenNodeOption {
	return func(options *OpenNodeOptions) {
		options.ForceInode = &inode
	}
}
