package psidsadapter

import (
	"context"
	"encoding/hex"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/dsadapter"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/psids"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/graphfs"
)

type DataStoreSuperBlock struct {
	graphfs.SuperBlockBase

	ds   datastore.Batching
	lsys linking.LinkSystem
	bmp  *SparseBitmapIndex

	mu   sync.RWMutex
	root *graphfs.CacheEntry
}

func NewDataStoreSuperBlock(ds datastore.Batching, uuid string) graphfs.SuperBlock {
	sb := &DataStoreSuperBlock{
		ds: ds,

		bmp: NewSparseBitmapIndex(),
	}

	dsa := &dsadapter.Adapter{
		Wrapped: sb.ds,

		EscapingFunc: func(s string) string {
			return "_cas/" + hex.EncodeToString([]byte(s))
		},
	}

	sb.lsys = cidlink.DefaultLinkSystem()
	sb.lsys.SetReadStorage(dsa)
	sb.lsys.SetWriteStorage(dsa)
	sb.lsys.TrustedStorage = true

	bmp, err := psids.Get(context.Background(), sb.ds, dsKeyBitmap)

	if err == nil {
		sb.bmp.LoadSnapshot(bmp)
	}

	sb.Init(sb, uuid, sb, sb)

	return sb
}

func (sb *DataStoreSuperBlock) AllocateINode() *graphfs.INode {
	id := int64(sb.bmp.Allocate())

	s := sb.bmp.Snapshot()

	if err := psids.Put(context.Background(), sb.ds, dsKeyBitmap, s); err != nil {
		panic(err)
	}

	return sb.MakeINode(id)
}

func (sb *DataStoreSuperBlock) DestroyINode(ino *graphfs.INode) {
	sb.bmp.Free(uint64(ino.ID()))
}

func (sb *DataStoreSuperBlock) MakeINode(id int64) *graphfs.INode {
	return graphfs.AllocateInode(sb, id)
}

func (sb *DataStoreSuperBlock) GetRoot(ctx context.Context) (*graphfs.CacheEntry, error) {
	if sb.root != nil {
		return sb.root, nil
	}

	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.root == nil {
		rootInode := sb.MakeINode(0)

		sb.root = graphfs.AllocCacheEntryRoot(sb)
		sb.root.Add(rootInode)
		sb.root = sb.root.Get()
	}

	return sb.root, nil
}

func (sb *DataStoreSuperBlock) Create(ctx context.Context, self *graphfs.CacheEntry, options graphfs.OpenNodeOptions) (graphfs.NodeHandle, error) {
	ino := self.Inode()

	if ino == nil && options.Flags&graphfs.OpenNodeFlagsCreate == 0 {
		return nil, psi.ErrNodeNotFound
	}

	if ino == nil {
		if p := self.Parent(); p != nil {
			k := dsKeyNodeEdge(p.Inode().ID(), self.Name())
			fe, err := psids.Get(ctx, sb.ds, k)

			if err == nil {
				ino = sb.MakeINode(fe.ToIndex)
			}
		}

		if ino == nil {
			ino = sb.AllocateINode()
		}

		self.Instantiate(ino)
	}

	return graphfs.NewNodeHandle(ino, self, options), nil
}

func (sb *DataStoreSuperBlock) Lookup(ctx context.Context, self *graphfs.INode, dentry *graphfs.CacheEntry) (*graphfs.CacheEntry, error) {
	k := dsKeyNodeEdge(self.ID(), dentry.Name())
	fe, err := psids.Get(ctx, sb.ds, k)

	if fe == nil || err == datastore.ErrNotFound {
		dentry.Add(nil)
	} else if err != nil {
		return nil, err
	} else {
		ino := sb.MakeINode(fe.ToIndex)

		dentry.Add(ino)
	}

	return dentry, nil
}

func (sb *DataStoreSuperBlock) Read(ctx context.Context, nh graphfs.NodeHandle) (*graphfs.SerializedNode, error) {
	k := dsKeyNodeData(nh.Inode().ID())

	return psids.Get(ctx, sb.ds, k)
}

func (sb *DataStoreSuperBlock) Write(ctx context.Context, nh graphfs.NodeHandle, fe *graphfs.SerializedNode) (ipld.Link, error) {
	writer := GetBatchWriter(ctx, sb.ds)
	k := dsKeyNodeData(nh.Inode().ID())

	if err := psids.Put(ctx, writer, k, fe); err != nil {
		return nil, err
	}

	return nil, nil
}

func (sb *DataStoreSuperBlock) SetEdge(ctx context.Context, nh graphfs.NodeHandle, edge *graphfs.SerializedEdge) error {
	writer := GetBatchWriter(ctx, sb.ds)
	k := dsKeyNodeEdge(nh.Inode().ID(), edge.Key)

	return psids.Put(ctx, writer, k, edge)
}

func (sb *DataStoreSuperBlock) RemoveEdge(ctx context.Context, nh graphfs.NodeHandle, key psi.EdgeKey) error {
	writer := GetBatchWriter(ctx, sb.ds)
	k := dsKeyNodeEdge(nh.Inode().ID(), key)

	return psids.Delete(ctx, writer, k)
}

func (sb *DataStoreSuperBlock) ReadEdge(ctx context.Context, nh graphfs.NodeHandle, key psi.EdgeKey) (*graphfs.SerializedEdge, error) {
	k := dsKeyNodeEdge(nh.Inode().ID(), key)

	return psids.Get(ctx, sb.ds, k)
}

func (sb *DataStoreSuperBlock) ReadEdges(ctx context.Context, nh graphfs.NodeHandle) (iterators.Iterator[*graphfs.SerializedEdge], error) {
	k := dsKeyEdgePrefix(nh.Inode().ID())

	return psids.List(ctx, sb.ds, k)
}
