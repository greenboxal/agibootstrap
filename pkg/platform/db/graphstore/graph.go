package graphstore

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/linking"
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/psids"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/graphfs"
	"github.com/greenboxal/agibootstrap/psidb/online"
	"github.com/greenboxal/agibootstrap/psidb/providers/psidsadapter"
)

type nodeUpdateRequest struct {
	Fence  uint64
	Node   psi.Node
	Frozen *psi.FrozenNode
	Edges  []*psi.FrozenEdge
	Link   ipld.Link
}

type IndexedGraph struct {
	logger *zap.SugaredLogger
	mu     sync.RWMutex

	root psi.UniqueNode

	ds         datastore.Batching
	journal    *graphfs.Journal
	checkpoint graphfs.Checkpoint

	bmp *SparseBitmapIndex
	vg  *graphfs.VirtualGraph
	lg  *online.LiveGraph

	proc            goprocess.Process
	nodeUpdateQueue chan nodeUpdateRequest
	closeCh         chan struct{}

	listeners []*listenerSlot
}

func NewIndexedGraph(ds datastore.Batching, walPath string, root psi.UniqueNode) (*IndexedGraph, error) {
	if err := os.MkdirAll(walPath, 0755); err != nil {
		return nil, err
	}

	journal, err := graphfs.OpenJournal(walPath)

	if err != nil {
		return nil, err
	}

	checkpoint, err := graphfs.OpenFileCheckpoint(path.Join(walPath, "ckpt"))

	if err != nil {
		return nil, err
	}

	sb := psidsadapter.NewDataStoreSuperBlock(ds, root.UUID())

	spb := graphfs.SuperBlockProvider(func(ctx context.Context, uuid string) (graphfs.SuperBlock, error) {
		return sb, nil
	})

	vg := graphfs.NewVirtualGraph(spb, journal, checkpoint)

	g := &IndexedGraph{
		logger: logging.GetLogger("graphstore"),

		bmp: NewSparseBitmapIndex(),

		root:       root,
		ds:         ds,
		vg:         vg,
		journal:    journal,
		checkpoint: checkpoint,

		closeCh:         make(chan struct{}),
		nodeUpdateQueue: make(chan nodeUpdateRequest, 8192),
	}

	g.lg = online.NewLiveGraph(g.vg)

	g.Add(root)

	g.proc = goprocess.Go(g.run)

	return g, nil
}

func (g *IndexedGraph) Root() psi.UniqueNode            { return g.root }
func (g *IndexedGraph) Store() *Store                   { return nil }
func (g *IndexedGraph) LinkSystem() *linking.LinkSystem { return nil }
func (g *IndexedGraph) DataStore() datastore.Batching   { return g.ds }

func (g *IndexedGraph) NextNodeID() int64      { return int64(g.bmp.Allocate()) }
func (g *IndexedGraph) NextEdgeID() psi.EdgeID { return psi.EdgeID(g.bmp.Allocate()) }

func (g *IndexedGraph) ResolveNode(ctx context.Context, path psi.Path) (n psi.Node, err error) {
	return g.lg.ResolveNode(ctx, path)

	/*needsTraversal := false

	if path.IsRelative() {
		path = g.root.CanonicalPath().Join(path)
	}

	for path := path; !path.IsEmpty(); path = path.Parent() {
		entry := g.getCacheEntry(path, true)

		if err := entry.Load(ctx); err != nil {
			if err == psi.ErrNodeNotFound {
				break
			}

			return nil, err
		}

		n = entry.Node()

		if n != nil {
			break
		}

		needsTraversal = true
	}

	if needsTraversal {
		rel, err := path.RelativeTo(n.CanonicalPath())

		if err != nil {
			return nil, err
		}

		return psi.ResolvePath(ctx, n, rel)
	}

	if n == nil {
		return nil, psi.ErrNodeNotFound
	}

	return n, nil*/
}

func (g *IndexedGraph) ListNodeEdges(ctx context.Context, path psi.Path) (result []*psi.FrozenEdge, err error) {
	if path.IsRelative() {
		return nil, fmt.Errorf("path must be absolute")
	}

	edges, err := g.vg.ReadEdges(ctx, path)

	if err != nil {
		return nil, err
	}

	result = iterators.ToSlice(iterators.Map(edges, func(edge *graphfs.SerializedEdge) *psi.FrozenEdge {
		return &psi.FrozenEdge{
			Key:     edge.Key,
			ToPath:  edge.ToPath,
			ToIndex: edge.ToIndex,
		}
	}))

	return result, nil
}

func (g *IndexedGraph) Add(n psi.Node) {
	if n.PsiNode() == nil {
		panic("node is not initialized")
	}

	_, err := g.lg.Add(context.Background(), n)

	if err != nil {
		panic(err)
	}

	n.PsiNodeBase().AttachToGraph(g)
}

func (g *IndexedGraph) Remove(n psi.Node) {
	g.lg.Remove(context.Background(), n)

}

func (g *IndexedGraph) SetEdge(e psi.Edge) {
}

func (g *IndexedGraph) UnsetEdge(self psi.Edge) {
}

func (g *IndexedGraph) RefreshNode(ctx context.Context, n psi.Node) error {
	snap, err := g.lg.Add(ctx, n)

	if err != nil {
		return err
	}

	return snap.Load(ctx)
}

func (g *IndexedGraph) CommitNode(ctx context.Context, node psi.Node) (ipld.Link, error) {
	err := g.lg.CommitNode(ctx, node)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (g *IndexedGraph) Commit(ctx context.Context) error {
	if err := g.root.Update(ctx); err != nil {
		return err
	}

	if _, err := g.CommitNode(ctx, g.root); err != nil {
		return err
	}

	batch, err := g.ds.Batch(ctx)

	if err != nil {
		return err
	}

	if err := g.flushRoot(ctx, batch); err != nil {
		return err
	}

	if err := g.flushBitmap(ctx, batch); err != nil {
		return err
	}

	return batch.Commit(ctx)
}

func (g *IndexedGraph) flushRoot(ctx context.Context, batch datastore.Batch) error {
	if err := psids.Put(ctx, batch, dsKeyRootPath, g.root.CanonicalPath()); err != nil {
		return err
	}

	if err := psids.Put(ctx, batch, dsKeyRootUuid, g.root.UUID()); err != nil {
		return err
	}

	return nil
}

func (g *IndexedGraph) loadBitmap(ctx context.Context) error {
	snapshot, err := psids.Get(ctx, g.ds, dsKeyBitmap)

	if err != nil {
		if errors.Is(err, psi.ErrNodeNotFound) {
			return nil
		}

		return err
	}

	g.bmp.LoadSnapshot(snapshot)

	return nil
}

func (g *IndexedGraph) flushBitmap(ctx context.Context, batch datastore.Batch) error {
	serialized := g.bmp.Snapshot()

	return psids.Put(ctx, g.ds, dsKeyBitmap, serialized)
}

func (g *IndexedGraph) OnNodeInvalidated(n psi.Node) {
}

func (g *IndexedGraph) OnNodeUpdated(n psi.Node) {
	if _, err := g.CommitNode(context.Background(), n); err != nil {
		panic(err)
	}
}

func (g *IndexedGraph) run(proc goprocess.Process) {
	ctx := goprocessctx.OnClosingContext(proc)

	defer func() {
		var wg sync.WaitGroup

		for _, listener := range g.listeners {
			wg.Add(1)

			go func(listener *listenerSlot) {
				if err := listener.proc.Close(); err != nil {
					g.logger.Error(err)
				}

				wg.Done()
			}(listener)
		}

		wg.Wait()
	}()

	if err := g.loadBitmap(ctx); err != nil {
		panic(err)
	}

	for {
		select {
		case <-ctx.Done():
			return

		case <-g.closeCh:
			for {
				select {
				case item := <-g.nodeUpdateQueue:
					if err := g.processQueueItem(ctx, item); err != nil {
						g.logger.Error(err)
					}

				default:
					return
				}
			}

		case item := <-g.nodeUpdateQueue:
			if err := g.processQueueItem(ctx, item); err != nil {
				g.logger.Error(err)
			}
		}
	}
}

func (g *IndexedGraph) processQueueItem(ctx context.Context, item nodeUpdateRequest) error {
	err := psids.Put(ctx, g.ds, dsKeyLastFence, item.Fence)

	if err != nil {
		return err
	}

	g.notifyNodeUpdated(ctx, item.Node)

	return nil
}

func (g *IndexedGraph) AddListener(l IndexedGraphListener) {
	g.mu.Lock()
	defer g.mu.Unlock()

	index := slices.IndexFunc(g.listeners, func(s *listenerSlot) bool {
		return s.listener == l
	})

	if index != -1 {
		return
	}

	s := &listenerSlot{
		g:        g,
		listener: l,
		queue:    make(chan psi.Node, 128),
	}

	s.proc = goprocess.SpawnChild(g.proc, s.run)

	g.listeners = append(g.listeners, s)
}

func (g *IndexedGraph) RemoveListener(l IndexedGraphListener) {
	g.mu.Lock()
	defer g.mu.Unlock()

	index := slices.IndexFunc(g.listeners, func(s *listenerSlot) bool {
		return s.listener == l
	})

	if index == -1 {
		return
	}

	s := g.listeners[index]

	g.listeners = slices.Delete(g.listeners, index, index+1)

	if err := s.proc.Close(); err != nil {
		panic(err)
	}
}

func (g *IndexedGraph) Shutdown(ctx context.Context) error {
	if g.closeCh != nil {
		close(g.closeCh)
		g.closeCh = nil
	}

	if err := g.Commit(ctx); err != nil {
		return err
	}

	if err := g.proc.Close(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.journal != nil {
		if err := g.journal.Close(); err != nil {
			return err
		}

		g.journal = nil
	}

	if g.checkpoint != nil {
		if err := g.checkpoint.Close(); err != nil {
			return err
		}

		g.checkpoint = nil
	}

	if g.vg != nil {
		if err := g.vg.Close(ctx); err != nil {
			return err
		}

		g.vg = nil
	}

	return nil
}

func (g *IndexedGraph) notifyNodeUpdated(ctx context.Context, node psi.Node) {
	g.dispatchListeners(node)
}

func (g *IndexedGraph) dispatchListeners(node psi.Node) {
	if node == nil {
		return
	}

	for _, l := range g.listeners {
		if l.queue != nil {
			l.queue <- node
		}
	}
}

func (g *IndexedGraph) LiveGraph() *online.LiveGraph { return g.lg }
