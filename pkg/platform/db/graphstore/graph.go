package graphstore

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

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

type listenerSlot struct {
	listener IndexedGraphListener
	queue    chan psi.Node
	proc     goprocess.Process
}

type IndexedGraphListener interface {
	OnNodeUpdated(node psi.Node)
}

type IndexedGraph struct {
	psi.BaseGraph

	logger *zap.SugaredLogger
	mu     sync.RWMutex

	root psi.UniqueNode

	ds    datastore.Batching
	store *Store
	wal   *WriteAheadLog

	nodeCache map[psi.NodeID]*cachedNode

	proc            goprocess.Process
	closeCh         chan struct{}
	nodeUpdateQueue chan nodeUpdateRequest

	listeners []*listenerSlot
}

func NewIndexedGraph(ds datastore.Batching, walPath string, root psi.UniqueNode) (*IndexedGraph, error) {
	if err := os.MkdirAll(walPath, 0755); err != nil {
		return nil, err
	}

	wal, err := NewWriteAheadLog(walPath)

	if err != nil {
		return nil, err
	}

	store := NewStore(ds, wal, root)

	g := &IndexedGraph{
		logger: logging.GetLogger("graphstore"),

		ds:    ds,
		wal:   wal,
		store: store,
		root:  root,

		nodeCache: map[psi.NodeID]*cachedNode{},

		closeCh:         make(chan struct{}),
		nodeUpdateQueue: make(chan nodeUpdateRequest, 256),
	}

	g.Init(g)

	g.proc = goprocess.Go(g.run)

	return g, nil
}

func (g *IndexedGraph) Root() psi.UniqueNode { return g.root }
func (g *IndexedGraph) Store() *Store        { return g.store }

func (g *IndexedGraph) Commit(ctx context.Context) error {
	if err := g.root.Update(ctx); err != nil {
		return err
	}

	if _, err := g.CommitNode(ctx, g.root); err != nil {
		return err
	}

	return nil
}

func (g *IndexedGraph) NewTransaction(root psi.Node) Transaction {
	return newGraphTx(g, root)
}

func (g *IndexedGraph) ResolveNode(ctx context.Context, path psi.Path) (n psi.Node, err error) {
	entry := g.getCacheEntry(path, true)

	if err := entry.Load(ctx); err != nil {
		return nil, err
	}

	n = entry.Node()

	if n == nil {
		return nil, psi.ErrNodeNotFound
	}

	return n, nil
}

func (g *IndexedGraph) ListNodeChildren(ctx context.Context, path psi.Path) (result []psi.Path, err error) {
	edges, err := g.store.ListNodeEdges(ctx, g.root.UUID(), path)

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
	entry := g.getCacheEntry(n.CanonicalPath(), true)

	doAdd := func() bool {
		entry.mu.Lock()
		defer entry.mu.Unlock()

		if entry.node == n {
			return false
		}

		if entry.node != nil {
			//panic("node already exists in graph")
			return false
		}

		entry.node = n

		return true
	}

	if doAdd() {
		g.BaseGraph.Add(n)

		if _, err := g.CommitNode(context.Background(), n); err != nil {
			panic(err)
		}
	}
}

func (g *IndexedGraph) Remove(n psi.Node) {
	entry := g.getCacheEntry(n.CanonicalPath(), false)

	if entry == nil {
		return
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	entry.node = nil
	entry.frozen = nil

	delete(g.nodeCache, n.CanonicalPath().String())

	if entry.node != nil {
		g.BaseGraph.Remove(n)
	}

	if err := g.store.RemoveNode(context.Background(), n.CanonicalPath()); err != nil {
		panic(err)
	}
}

func (g *IndexedGraph) RefreshNode(ctx context.Context, n psi.Node) error {
	entry := g.getCacheEntry(n.CanonicalPath(), true)

	if entry.node == nil {
		return entry.updateFromMemory(ctx, n)
	} else if entry.node != n {
		return fmt.Errorf("conflict: node already exists in graph")
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

	if err := entry.updateFromMemory(ctx, node); err != nil {
		return nil, err
	}

	if err := entry.Commit(ctx, nil); err != nil {
		return nil, err
	}

	return entry.CommittedLink(), nil
}

func (g *IndexedGraph) OnNodeInvalidated(n psi.Node) {
}

func (g *IndexedGraph) OnNodeUpdated(n psi.Node) {
	if _, err := g.CommitNode(context.Background(), n); err != nil {
		panic(err)
	}
}

func (g *IndexedGraph) getCacheEntry(path psi.Path, create bool) *cachedNode {
	if create {
		g.mu.Lock()
		defer g.mu.Unlock()
	} else {
		g.mu.RLock()
		defer g.mu.RUnlock()
	}

	key := path.String()
	entry := g.nodeCache[key]

	if entry == nil && create {
		entry = &cachedNode{
			g: g,

			path: path,
		}

		g.nodeCache[key] = entry
	}

	return entry
}

func (g *IndexedGraph) run(proc goprocess.Process) {
	ctx := goprocessctx.OnClosingContext(proc)

	/*if err := g.recoverFromWal(ctx); err != nil {
		panic(err)
	}*/

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
	lastFenceData, err := g.ds.Get(ctx, lastFenceKey)

	if err != nil && err != datastore.ErrNotFound {
		return err
	}

	if len(lastFenceData) == 0 {
		return nil
	}

	lastFence, err := strconv.ParseUint(string(lastFenceData), 10, 64)

	if err != nil {
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
	err := g.ds.Put(ctx, lastFenceKey, []byte(strconv.FormatUint(item.Fence, 10)))

	if err != nil {
		return err
	}

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
