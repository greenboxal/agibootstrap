package localfs

import (
	"context"

	"github.com/greenboxal/agibootstrap/psidb/graphfs"
)

type SuperBlock struct {
	graphfs.SuperBlockBase

	root string
}

func NewSuperBlock(root string) graphfs.SuperBlock {
	sb := &SuperBlock{
		root: root,
	}

	sb.Init(sb, "", sb, sb)

	return sb
}

func (sb *SuperBlock) GetRoot(ctx context.Context) (*graphfs.CacheEntry, error) {
	//TODO implement me
	panic("implement me")
}

func (sb *SuperBlock) Create(ctx context.Context, self *graphfs.CacheEntry, options graphfs.OpenNodeOptions) (graphfs.NodeHandle, error) {
	//TODO implement me
	panic("implement me")
}

func (sb *SuperBlock) Lookup(ctx context.Context, self *graphfs.INode, dentry *graphfs.CacheEntry) (*graphfs.CacheEntry, error) {
	//TODO implement me
	panic("implement me")
}
