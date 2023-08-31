package rtv1

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

type SuperBlockAdapter struct {
	graphfs.SuperBlockBase

	uuid   string
	client NodeAPI
}

func NewSuperBlockAdapter(client NodeAPI, uuid string) *SuperBlockAdapter {
	sb := &SuperBlockAdapter{uuid: uuid, client: client}
	sb.Init(sb, uuid, sb, sb)

	return sb
}

func (sb *SuperBlockAdapter) UUID() string { return sb.uuid }

func (sb *SuperBlockAdapter) GetRoot(ctx context.Context) (*graphfs.CacheEntry, error) {
	//TODO implement me
	panic("implement me")
}

func (sb *SuperBlockAdapter) Create(ctx context.Context, self *graphfs.CacheEntry, options graphfs.OpenNodeOptions) (graphfs.NodeHandle, error) {
	//TODO implement me
	panic("implement me")
}

func (sb *SuperBlockAdapter) Lookup(ctx context.Context, self *graphfs.INode, dentry *graphfs.CacheEntry) (*graphfs.CacheEntry, error) {
	//TODO implement me
	panic("implement me")
}

func (sb *SuperBlockAdapter) Read(ctx context.Context, nh graphfs.NodeHandle) (*coreapi.SerializedNode, error) {

	//TODO implement me
	panic("implement me")
}

func (sb *SuperBlockAdapter) Write(ctx context.Context, nh graphfs.NodeHandle, fe *coreapi.SerializedNode) error {
	//TODO implement me
	panic("implement me")
}

func (sb *SuperBlockAdapter) SetEdge(ctx context.Context, nh graphfs.NodeHandle, edge *coreapi.SerializedEdge) error {
	//TODO implement me
	panic("implement me")
}

func (sb *SuperBlockAdapter) RemoveEdge(ctx context.Context, nh graphfs.NodeHandle, key psi.EdgeKey) error {
	//TODO implement me
	panic("implement me")
}

func (sb *SuperBlockAdapter) ReadEdge(ctx context.Context, nh graphfs.NodeHandle, key psi.EdgeKey) (*coreapi.SerializedEdge, error) {
	//TODO implement me
	panic("implement me")
}

func (sb *SuperBlockAdapter) ReadEdges(ctx context.Context, nh graphfs.NodeHandle) (iterators.Iterator[*coreapi.SerializedEdge], error) {
	//TODO implement me
	panic("implement me")
}

func (sb *SuperBlockAdapter) Close(ctx context.Context) error {
	return nil
}
