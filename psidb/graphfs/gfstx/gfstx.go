package gfstx

import (
	"context"

	"github.com/ipld/go-ipld-prime"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/graphfs"
)

type SuperBlock struct {
	graphfs.SuperBlockBase

	Super graphfs.SuperBlock
}

func NewSuperBlock(s graphfs.SuperBlock) *SuperBlock {
	sb := &SuperBlock{
		Super: s,
	}

	sb.Init(sb, s.UUID(), sb, sb)

	return sb
}

func (s *SuperBlock) GetRoot(ctx context.Context) (*graphfs.CacheEntry, error) {
	return s.Super.GetRoot(ctx)
}

func (s *SuperBlock) Create(ctx context.Context, self *graphfs.CacheEntry, options graphfs.OpenNodeOptions) (graphfs.NodeHandle, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SuperBlock) Lookup(ctx context.Context, self *graphfs.INode, dentry *graphfs.CacheEntry) (*graphfs.CacheEntry, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SuperBlock) Read(ctx context.Context, nh graphfs.NodeHandle) (*graphfs.SerializedNode, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SuperBlock) Write(ctx context.Context, nh graphfs.NodeHandle, fe *graphfs.SerializedNode) (ipld.Link, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SuperBlock) SetEdge(ctx context.Context, nh graphfs.NodeHandle, edge *graphfs.SerializedEdge) error {
	//TODO implement me
	panic("implement me")
}

func (s *SuperBlock) RemoveEdge(ctx context.Context, nh graphfs.NodeHandle, key psi.EdgeKey) error {
	//TODO implement me
	panic("implement me")
}

func (s *SuperBlock) ReadEdge(ctx context.Context, nh graphfs.NodeHandle, key psi.EdgeKey) (*graphfs.SerializedEdge, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SuperBlock) ReadEdges(ctx context.Context, nh graphfs.NodeHandle) (iterators.Iterator[*graphfs.SerializedEdge], error) {
	//TODO implement me
	panic("implement me")
}
