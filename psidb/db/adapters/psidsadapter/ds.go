package psidsadapter

import (
	"context"
	"encoding/hex"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/dsadapter"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/psids"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
	graphfs "github.com/greenboxal/agibootstrap/psidb/db/graphfs"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

var tracer = otel.Tracer("psidsadapter", trace.WithInstrumentationAttributes(
	semconv.ServiceName("psidb-graph"),
	semconv.DBSystemKey.String("badger"),
))

type DataStoreSuperBlock struct {
	graphfs.SuperBlockBase

	ds   datastore.Batching
	lsys linking.LinkSystem
	bmp  *SparseBitmapIndex

	mu   sync.RWMutex
	root *graphfs.CacheEntry

	shouldClose bool
}

func NewDataStoreSuperBlock(
	ctx context.Context,
	ds datastore.Batching,
	uuid string,
	shouldClose bool,
) (graphfs.SuperBlock, error) {
	sb := &DataStoreSuperBlock{
		ds:          ds,
		shouldClose: shouldClose,

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

	sb.Init(sb, uuid, sb, sb)

	if err := sb.LoadBitmap(ctx); err != nil {
		return nil, err
	}

	return sb, nil
}

func (sb *DataStoreSuperBlock) AllocateINode(ctx context.Context) *graphfs.INode {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	return sb.makeInodeLocked(-1)
}

func (sb *DataStoreSuperBlock) DestroyInode(ctx context.Context, ino *graphfs.INode) error {
	return nil
}

func (sb *DataStoreSuperBlock) DirtyInode(ctx context.Context, ino *graphfs.INode) error {
	return nil
}

func (sb *DataStoreSuperBlock) WriteInode(ctx context.Context, ino *graphfs.INode) error {
	return nil
}

func (sb *DataStoreSuperBlock) DropInode(ctx context.Context, ino *graphfs.INode) error {
	sb.bmp.Free(uint64(ino.ID()))

	return nil
}

func (sb *DataStoreSuperBlock) MakeInode(id int64) *graphfs.INode {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	return sb.makeInodeLocked(id)
}

func (sb *DataStoreSuperBlock) makeInodeLocked(id int64) *graphfs.INode {
	if id == -1 {
		id = int64(sb.bmp.Allocate())
	}

	ino := graphfs.AllocateInode(sb, id)

	return ino
}

func (sb *DataStoreSuperBlock) GetRoot(ctx context.Context) (*graphfs.CacheEntry, error) {
	if sb.root != nil {
		return sb.root, nil
	}

	ctx, span := tracer.Start(ctx, "DataStoreSuperBlock.GetRoot")
	defer span.End()

	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.root == nil {
		rootInode := sb.makeInodeLocked(0)

		sb.root = graphfs.AllocCacheEntryRoot(sb)
		sb.root.Add(rootInode)
		sb.root = sb.root.Get()
	}

	//logger.Debugf("GetRoot")

	return sb.root, nil
}

func (sb *DataStoreSuperBlock) Create(ctx context.Context, self *graphfs.CacheEntry, options graphfs.OpenNodeOptions) error {
	ctx, span := tracer.Start(ctx, "DataStoreSuperBlock.Create")
	defer span.End()

	var ino *graphfs.INode

	if options.ForceInode != nil && *options.ForceInode >= 0 {
		allocated := sb.bmp.MarkAllocated(uint64(*options.ForceInode))

		ino = sb.MakeInode(int64(allocated))
	} else {
		ino = sb.AllocateINode(ctx)
	}

	if self.Parent() != nil {
		k := dsKeyNodeEdge(self.Parent().Inode().ID(), self.Name())

		if err := psids.Put(ctx, sb.ds, k, &coreapi.SerializedEdge{
			Flags:   coreapi.EdgeFlagRegular,
			Key:     self.Name().AsEdgeKey(),
			ToIndex: ino.ID(),
			ToPath:  self.Path(),
		}); err != nil {
			return err
		}
	}

	self.Instantiate(ino)

	//logger.Debugw("Create", "ino", ino.ID(), "path", self.Path().String())

	return nil
}

func (sb *DataStoreSuperBlock) Lookup(ctx context.Context, self *graphfs.INode, dentry *graphfs.CacheEntry) (*graphfs.CacheEntry, error) {
	if self == nil {
		return nil, nil
	}

	ctx, span := tracer.Start(ctx, "DataStoreSuperBlock.Lookup")
	defer span.End()

	k := dsKeyNodeEdge(self.ID(), dentry.Name())
	fe, err := psids.Get(ctx, sb.ds, k)

	if err == nil && fe != nil {
		ino := sb.MakeInode(fe.ToIndex)

		dentry.Add(ino)

		//logger.Debugw("Lookup OK", "ino", ino.ID(), "path", dentry.Path().String())
	} else if !errors.Is(err, psi.ErrNodeNotFound) {
		//logger.Debugw("Lookup FAIL", "path", dentry.Path().String())
		return nil, err
	}

	return dentry, nil
}

func (sb *DataStoreSuperBlock) Unlink(ctx context.Context, self *graphfs.INode, dentry *graphfs.CacheEntry) error {
	//TODO implement me
	panic("implement me")
}

func (sb *DataStoreSuperBlock) Read(ctx context.Context, nh graphfs.NodeHandle) (*coreapi.SerializedNode, error) {
	ctx, span := tracer.Start(ctx, "DataStoreSuperBlock.Read")
	defer span.End()

	k := dsKeyNodeData(nh.Inode().ID())

	//logger.Debugw("Read", "k", k.String())

	return psids.Get(ctx, sb.ds, k)
}

func (sb *DataStoreSuperBlock) Write(ctx context.Context, nh graphfs.NodeHandle, fe *coreapi.SerializedNode) error {
	ctx, span := tracer.Start(ctx, "DataStoreSuperBlock.Write")
	defer span.End()

	writer := GetBatchWriter(ctx, sb.ds)
	k := dsKeyNodeData(nh.Inode().ID())

	if err := psids.Put(ctx, writer, k, fe); err != nil {
		return err
	}

	//logger.Debugw("Write", "k", k.String())

	return nil
}

func (sb *DataStoreSuperBlock) SetEdge(ctx context.Context, nh graphfs.NodeHandle, edge *coreapi.SerializedEdge) error {
	ctx, span := tracer.Start(ctx, "DataStoreSuperBlock.SetEdge")
	defer span.End()

	writer := GetBatchWriter(ctx, sb.ds)
	k := dsKeyNodeEdge(nh.Inode().ID(), edge.Key)

	//logger.Debugw("SetEdge", "k", k.String())

	return psids.Put(ctx, writer, k, edge)
}

func (sb *DataStoreSuperBlock) RemoveEdge(ctx context.Context, nh graphfs.NodeHandle, key psi.EdgeKey) error {
	ctx, span := tracer.Start(ctx, "DataStoreSuperBlock.RemoveEdge")
	defer span.End()

	writer := GetBatchWriter(ctx, sb.ds)
	k := dsKeyNodeEdge(nh.Inode().ID(), key)

	//logger.Debugw("RemoveEdge", "k", k.String())

	return psids.Delete(ctx, writer, k)
}

func (sb *DataStoreSuperBlock) ReadEdge(ctx context.Context, nh graphfs.NodeHandle, key psi.EdgeKey) (*coreapi.SerializedEdge, error) {
	ctx, span := tracer.Start(ctx, "DataStoreSuperBlock.ReadEdge")
	defer span.End()

	k := dsKeyNodeEdge(nh.Inode().ID(), key)

	//logger.Debugw("ReadEdge", "k", k.String())

	return psids.Get(ctx, sb.ds, k)
}

func (sb *DataStoreSuperBlock) ReadEdges(ctx context.Context, nh graphfs.NodeHandle) (iterators.Iterator[*coreapi.SerializedEdge], error) {
	ctx, span := tracer.Start(ctx, "DataStoreSuperBlock.ReadEdges")
	defer span.End()

	k := dsKeyEdgePrefix(nh.Inode().ID())

	//logger.Debugw("ReadEdges", "k", k.String())

	return psids.List(ctx, sb.ds, k)
}

func (sb *DataStoreSuperBlock) LoadBitmap(ctx context.Context) error {
	bmp, err := psids.Get(ctx, sb.ds, dsKeyBitmap)

	if err == nil {
		sb.bmp.LoadSnapshot(bmp)
	}

	return nil
}

func (sb *DataStoreSuperBlock) Flush(ctx context.Context) error {
	s := sb.bmp.Snapshot()

	if err := psids.Put(ctx, sb.ds, dsKeyBitmap, s); err != nil {
		panic(err)
	}

	return nil
}

func (sb *DataStoreSuperBlock) Close(ctx context.Context) error {
	if err := sb.Flush(ctx); err != nil {
		return err
	}

	if !sb.shouldClose {
		return nil
	}

	return sb.ds.Close()
}
