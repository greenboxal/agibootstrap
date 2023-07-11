package graphstore

import (
	"context"
	"fmt"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/psids"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/psi"
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

	ds    datastore.Batching
	store *Store
	wal   *WriteAheadLog

	bmp *SparseBitmapIndex

	nodeIdMap map[psi.NodeID]*cachedNode

	proc            goprocess.Process
	closeCh         chan struct{}
	nodeUpdateQueue chan nodeUpdateRequest

	listeners []*listenerSlot
}

func NewIndexedGraph(ds datastore.Batching, walPath string, root psi.UniqueNode) (*IndexedGraph, error) {
	wal, err := NewWriteAheadLog(walPath)

	if err != nil {
		return nil, err
	}

	store := NewStore(ds, wal, root)

	g := &IndexedGraph{
		logger: logging.GetLogger("graphstore"),

		bmp: NewSparseBitmapIndex(),

		ds:    ds,
		wal:   wal,
		store: store,
		root:  root,

		nodeIdMap: map[psi.NodeID]*cachedNode{},

		closeCh:         make(chan struct{}),
		nodeUpdateQueue: make(chan nodeUpdateRequest, 256),
	}

	g.Add(root)

	g.proc = goprocess.Go(g.run)

	return g, nil
}

func (g *IndexedGraph) Root() psi.UniqueNode            { return g.root }
func (g *IndexedGraph) Store() *Store                   { return g.store }
func (g *IndexedGraph) LinkSystem() *linking.LinkSystem { return &g.store.lsys }
func (g *IndexedGraph) DataStore() datastore.Batching   { return g.store.ds }

func (g *IndexedGraph) NextNodeID() int64      { return int64(g.bmp.Allocate()) }
func (g *IndexedGraph) NextEdgeID() psi.EdgeID { return psi.EdgeID(g.bmp.Allocate()) }

func (g *IndexedGraph) ResolveNode(ctx context.Context, path psi.Path) (n psi.Node, err error) {
	needsTraversal := false

	if path.IsRelative() {
		path = g.root.CanonicalPath().Join(path)
	}

	for path := path; !path.IsEmpty(); path = path.Parent() {
		entry := g.getCacheEntry(path, true)

		if err := entry.Load(ctx); err != nil {
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

	return n, nil
}

func (g *IndexedGraph) ResolveEdge(ctx context.Context, e *psi.FrozenEdge) (psi.Node, error) {
	if e.ToPath != nil {
		return g.ResolveNode(ctx, *e.ToPath)
	}

	if e.ToLink != nil {
		fn, err := g.store.GetNodeByCid(ctx, *e.ToLink)

		if err != nil {
			return nil, err
		}

		n, err := g.LoadNode(ctx, fn)

		if err != nil {
			return nil, err
		}

		return n, nil
	}

	return nil, fmt.Errorf("edge has no ToPath or ToLink")
}

func (g *IndexedGraph) ListNodeChildren(ctx context.Context, path psi.Path) (result []psi.Path, err error) {
	if path.IsRelative() {
		return nil, fmt.Errorf("path must be absolute")
	}

	edges, err := g.store.ListNodeEdges(ctx, path)

	if err != nil {
		return nil, err
	}

	for edges.Next() {
		fe := edges.Value()

		if fe.Key.Kind != psi.EdgeKindChild {
			continue
		}

		p := fe.FromPath.Child(fe.Key.AsPathElement())

		result = append(result, p)
	}

	return result, nil
}

func (g *IndexedGraph) Add(n psi.Node) {
	if n.PsiNode() == nil {
		panic("node is not initialized")
	}

	snap := n.PsiNodeBase().GetSnapshot()

	if snap != nil {
		return
	}

	entry := g.getCacheEntry(n.CanonicalPath(), true)

	if err := entry.updateNode(n); err != nil {
		panic(err)
	}

	n.PsiNodeBase().AttachToGraph(g)
}

func (g *IndexedGraph) Remove(n psi.Node) {
	entry := g.getCacheEntry(n.CanonicalPath(), false)

	if entry == nil {
		return
	}

	if err := entry.Remove(context.Background(), n); err != nil {
		panic(err)
	}
}

func (g *IndexedGraph) SetEdge(e psi.Edge) {
}

func (g *IndexedGraph) UnsetEdge(self psi.Edge) {
}

func (g *IndexedGraph) RefreshNode(ctx context.Context, n psi.Node) error {
	entry := g.getCacheEntry(n.CanonicalPath(), true)

	if entry.node == nil {
		if err := entry.updateNode(n); err != nil {
			return err
		}
	} else if entry.node != n {
		return fmt.Errorf("node already exists in graph: %s", n.CanonicalPath())
	}

	if entry.frozen == nil {
		if err := entry.Preload(ctx); err != nil {
			return err
		}
	}

	return entry.Refresh(ctx)
}

func (g *IndexedGraph) LoadNode(ctx context.Context, fn *psi.FrozenNode) (psi.Node, error) {
	entry := g.getCacheEntry(fn.Path, true)

	if err := entry.Load(ctx); err != nil {
		return nil, err
	}

	return entry.Node(), nil
}

func (g *IndexedGraph) CommitNode(ctx context.Context, node psi.Node) (ipld.Link, error) {
	entry := g.getCacheEntry(node.CanonicalPath(), true)

	if err := entry.updateNode(node); err != nil {
		return nil, err
	}

	if err := entry.Commit(ctx, nil); err != nil {
		return nil, err
	}

	return entry.CommitLink(), nil
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

	if snap := g.root.PsiNodeBase().GetSnapshot(); snap != nil {
		if err := psids.Put(ctx, batch, dsKeyRootSnapshot, snap.CommitLink().(cidlink.Link)); err != nil {
			return err
		}
	}

	return nil
}

func (g *IndexedGraph) loadBitmap(ctx context.Context) error {
	snapshot, err := psids.Get(ctx, g.ds, dsKeyBitmap)

	if err != nil {
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

func (g *IndexedGraph) getCacheEntry(path psi.Path, create bool) *cachedNode {
	if path.IsRelative() {
		panic(fmt.Errorf("path must be absolute"))
	}

	if create {
		g.mu.Lock()
		defer g.mu.Unlock()
	} else {
		g.mu.RLock()
		defer g.mu.RUnlock()
	}

	key := path.String()
	entry := g.nodeIdMap[key]

	if entry == nil && create {
		entry = &cachedNode{
			g:    g,
			id:   -1,
			path: path,
		}

		g.nodeIdMap[key] = entry
	}

	return entry
}

func (g *IndexedGraph) run(proc goprocess.Process) {
	ctx := goprocessctx.OnClosingContext(proc)

	if err := g.recoverFromWal(ctx); err != nil {
		panic(err)
	}

	if err := g.loadBitmap(ctx); err != nil {

	}

	for {
		select {
		case <-g.closeCh:
			return

		case <-ctx.Done():
			return

		case item := <-g.nodeUpdateQueue:
			if err := g.processQueueItem(ctx, item); err != nil {
				g.logger.Error(err)
			}
		}
	}
}

func (g *IndexedGraph) recoverFromWal(ctx context.Context) error {
	lastFence, err := psids.Get(ctx, g.ds, dsKeyLastFence)

	if err != nil {
		if err == psi.ErrNodeNotFound {
			return nil
		}

		return err
	}

	for i := lastFence; i <= g.wal.LastRecordIndex(); i++ {
		rec, err := g.wal.ReadRecord(i)

		if err != nil {
			return err
		}

		if rec.Op != WalOpUpdateNode {
			continue
		}

		if rec.Payload == nil {
			continue
		}

		fnLink := cidlink.Link{Cid: *rec.Payload}
		fn, err := g.store.GetNodeByCid(ctx, fnLink)

		if err != nil {
			return err
		}

		edges := make([]*psi.FrozenEdge, len(fn.Edges))

		for i, edgeLink := range fn.Edges {
			edge, err := g.store.GetEdgeByCid(ctx, edgeLink)

			if err != nil {
				return err
			}

			edges[i] = edge
		}

		item := nodeUpdateRequest{
			Fence:  rec.Counter,
			Frozen: fn,
			Edges:  edges,
			Link:   fnLink,
		}

		if err := g.processQueueItem(ctx, item); err != nil {
			g.logger.Error(err)
		}
	}

	return nil
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
		listener: l,
		queue:    make(chan psi.Node, 128),
	}

	s.proc = goprocess.SpawnChild(g.proc, func(proc goprocess.Process) {
		for {
			select {
			case <-proc.Closing():
				return
			case n := <-s.queue:
				func() {
					defer func() {
						if r := recover(); r != nil {
							g.logger.Error("panic in graph listener", "err", r)
						}
					}()

					l.OnNodeUpdated(n)
				}()
			}
		}
	})

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
	if err := g.Commit(ctx); err != nil {
		return err
	}

	if g.proc != nil {
		close(g.closeCh)

		if err := g.proc.CloseAfterChildren(); err != nil {
			return err
		}

		g.proc = nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.wal != nil {
		if err := g.wal.Close(); err != nil {
			return err
		}

		g.wal = nil
	}

	return nil
}

func (g *IndexedGraph) notifyNodeUpdated(ctx context.Context, node psi.Node) {
	g.dispatchListeners(node)
}

func (g *IndexedGraph) dispatchListeners(node psi.Node) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, l := range g.listeners {
		if l.queue != nil {
			l.queue <- node
		}
	}
}
