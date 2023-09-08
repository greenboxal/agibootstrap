package remotesb

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/psidb/apis/rt/v1"
	"github.com/greenboxal/agibootstrap/psidb/client"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type RemoteSuperBlock struct {
	graphfs.SuperBlockBase

	mu sync.RWMutex

	uuid   string
	driver *client.Driver

	root *graphfs.CacheEntry
}

func NewRemoteSuperBlock(driver *client.Driver, uuid string) *RemoteSuperBlock {
	return &RemoteSuperBlock{
		driver: driver,
		uuid:   uuid,
	}
}

func (sb *RemoteSuperBlock) UUID() string {
	return sb.uuid
}

func (sb *RemoteSuperBlock) AllocateINode(ctx context.Context) *graphfs.INode {
	return graphfs.AllocateInode(sb, -1)
}

func (sb *RemoteSuperBlock) MakeInode(id int64) *graphfs.INode {
	return graphfs.AllocateInode(sb, id)
}

func (sb *RemoteSuperBlock) DestroyInode(ctx context.Context, ino *graphfs.INode) error {
	return coreapi.ErrUnsupportedOperation
}

func (sb *RemoteSuperBlock) DirtyInode(ctx context.Context, ino *graphfs.INode) error {
	return coreapi.ErrUnsupportedOperation
}

func (sb *RemoteSuperBlock) WriteInode(ctx context.Context, ino *graphfs.INode) error {
	return coreapi.ErrUnsupportedOperation
}

func (sb *RemoteSuperBlock) DropInode(ctx context.Context, ino *graphfs.INode) error {
	return coreapi.ErrUnsupportedOperation
}

func (sb *RemoteSuperBlock) GetRoot(ctx context.Context) (*graphfs.CacheEntry, error) {
	sb.mu.RLock()
	if root := sb.root; root != nil {
		sb.mu.RUnlock()
		return root, nil
	}
	sb.mu.RUnlock()

	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.root == nil {
		rootInode := graphfs.AllocateInode(sb, 0)

		sb.root = graphfs.AllocCacheEntryRoot(sb)
		sb.root.Add(rootInode)
		sb.root = sb.root.Get()
	}

	return sb.root, nil
}

func (sb *RemoteSuperBlock) Create(ctx context.Context, self *graphfs.CacheEntry, options graphfs.OpenNodeOptions) error {
	return coreapi.ErrUnsupportedOperation
}

func (sb *RemoteSuperBlock) Lookup(ctx context.Context, self *graphfs.INode, dentry *graphfs.CacheEntry) (*graphfs.CacheEntry, error) {
	res, err := sb.driver.LookupNode(ctx, &v1.LookupNodeRequest{
		Inode: self.ID(),
		Path:  dentry.Path(),
	})

	if err == nil && res.Data != nil {
		ino := sb.MakeInode(res.Data.Index)

		dentry.Add(ino)

	} else if !errors.Is(err, psi.ErrNodeNotFound) {
		return nil, err
	}

	return dentry, nil
}

func (sb *RemoteSuperBlock) Unlink(ctx context.Context, self *graphfs.INode, dentry *graphfs.CacheEntry) error {
	return coreapi.ErrUnsupportedOperation
}

func (sb *RemoteSuperBlock) Read(ctx context.Context, nh graphfs.NodeHandle) (*coreapi.SerializedNode, error) {
	res, err := sb.driver.ReadNode(ctx, &v1.ReadNodeRequest{
		Inode: nh.Inode().ID(),
		Path:  nh.Entry().Path(),
	})

	if err != nil {
		return nil, err
	}

	return res.Data, nil
}

func (sb *RemoteSuperBlock) Write(ctx context.Context, nh graphfs.NodeHandle, fe *coreapi.SerializedNode) error {
	return coreapi.ErrUnsupportedOperation
}

func (sb *RemoteSuperBlock) SetEdge(ctx context.Context, nh graphfs.NodeHandle, edge *coreapi.SerializedEdge) error {
	return coreapi.ErrUnsupportedOperation
}

func (sb *RemoteSuperBlock) RemoveEdge(ctx context.Context, nh graphfs.NodeHandle, key psi.EdgeKey) error {
	return coreapi.ErrUnsupportedOperation
}

func (sb *RemoteSuperBlock) ReadEdge(ctx context.Context, nh graphfs.NodeHandle, key psi.EdgeKey) (*coreapi.SerializedEdge, error) {
	res, err := sb.driver.ReadEdge(ctx, &v1.ReadEdgeRequest{
		ParentInode: nh.Inode().ID(),
		Path:        nh.Entry().Path(),
	})

	if err != nil {
		return nil, err
	}

	return res.Data, nil
}

func (sb *RemoteSuperBlock) ReadEdges(ctx context.Context, nh graphfs.NodeHandle) (iterators.Iterator[*coreapi.SerializedEdge], error) {
	res, err := sb.driver.ReadEdgesStream(ctx, &v1.ReadEdgesRequest{
		Inode: nh.Inode().ID(),
		Path:  nh.Entry().Path(),
	})

	if err != nil {
		return nil, err
	}

	it := iterators.FlatMap(res, func(t *v1.ReadEdgesResponse) iterators.Iterator[*coreapi.SerializedEdge] {
		return iterators.FromSlice(t.Data)
	})

	return it, nil
}

func (sb *RemoteSuperBlock) Close(ctx context.Context) error {
	return nil
}
