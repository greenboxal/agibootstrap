package graphstore

import (
	"context"
	"sync"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime"
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type nodeUpdateRequest struct {
	Node    psi.Node
	Version int64
}

type cachedNode struct {
	mu sync.Mutex

	uuid   psi.NodeID
	frozen *psi.FrozenNode
	node   psi.Node
}

type IndexedGraphListener interface {
	OnNodeUpdated(node psi.Node)
}

type listenerSlot struct {
	listener IndexedGraphListener
	queue    chan psi.Node
	proc     goprocess.Process
}

type IndexedGraph struct {
	psi.BaseGraph

	logger *zap.SugaredLogger
	mu     sync.RWMutex

	store *Store
	root  psi.Node

	nodeCache map[psi.NodeID]*cachedNode

	proc            goprocess.Process
	nodeUpdateQueue chan nodeUpdateRequest

	listeners []*listenerSlot
}

func NewIndexedGraph(ds datastore.Batching, root psi.Node) *IndexedGraph {
	store := NewStore(ds)

	g := &IndexedGraph{
		logger: logging.GetLogger("graphstore"),

		root:  root,
		store: store,

		nodeCache: map[psi.NodeID]*cachedNode{},

		nodeUpdateQueue: make(chan nodeUpdateRequest, 256),
	}

	g.Init(g)

	g.proc = goprocess.Go(g.run)

	return g
}

func (g *IndexedGraph) getCacheEntry(id psi.NodeID, create bool) *cachedNode {
	if create {
		g.mu.Lock()
		defer g.mu.Unlock()
	} else {
		g.mu.RLock()
		defer g.mu.RUnlock()
	}

	entry := g.nodeCache[id]

	if entry == nil && create {
		entry = &cachedNode{
			uuid: id,
		}

		g.nodeCache[id] = entry
	}

	return entry
}

func (g *IndexedGraph) loadCacheEntry(ctx context.Context, entry *cachedNode) error {
	entry.mu.Lock()
	defer entry.mu.Unlock()

	if entry.node != nil {
		return nil
	}

	if entry.frozen == nil {
		frozen, err := g.store.GetNodeByID(ctx, entry.uuid, -1)

		if err != nil {
			return err
		}

		entry.frozen = frozen
	}

	if entry.node == nil {
		node, err := g.store.LoadNode(ctx, entry.frozen)

		if err != nil {
			return err
		}

		entry.node = node
	}

	return nil
}

func (g *IndexedGraph) Add(n psi.Node) {
	entry := g.getCacheEntry(n.UUID(), true)

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

		frozen, err := g.store.UpsertNode(context.Background(), n)

		if err != nil {
			panic(err)
		}

		entry.frozen = frozen
		entry.node = n

		return true
	}

	if doAdd() {
		g.BaseGraph.Add(n)

		g.nodeUpdateQueue <- nodeUpdateRequest{
			Node:    n,
			Version: n.PsiNodeVersion(),
		}
	}
}

func (g *IndexedGraph) Remove(n psi.Node) {
	entry := g.getCacheEntry(n.UUID(), false)

	if entry == nil {
		return
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	entry.node = nil
	entry.frozen = nil

	delete(g.nodeCache, n.UUID())

	if entry.node != nil {
		g.BaseGraph.Remove(n)
	}

	if err := g.store.RemoveNode(context.Background(), n.CanonicalPath()); err != nil {
		panic(err)
	}
}

func (g *IndexedGraph) ResolveNode(path psi.Path) (n psi.Node, err error) {
	return psi.ResolvePath(g.root, path)
}

func (g *IndexedGraph) GetNodeByID(id psi.NodeID) (psi.Node, error) {
	entry := g.getCacheEntry(id, true)

	if err := g.loadCacheEntry(context.Background(), entry); err != nil {
		return nil, err
	}

	if entry.node == nil {
		return nil, psi.ErrNodeNotFound
	}

	return entry.node, nil
}

func (g *IndexedGraph) GetNodeChildren(path psi.Path) (result []psi.Path, err error) {
	var n psi.Node

	if path.Root() != nil {
		n, err = psi.ResolvePath(path.Root(), path)

		if err != nil {
			return nil, err
		}
	} else {
		n, err = g.ResolveNode(path)

		if err != nil {
			return nil, err
		}
	}

	return lo.Map(n.Children(), func(c psi.Node, _ int) psi.Path {
		return c.CanonicalPath()
	}), nil
}

func (g *IndexedGraph) CommitNode(ctx context.Context, node psi.Node) (ipld.Link, error) {
	if _, err := g.store.UpsertNode(ctx, node); err != nil {
		return nil, err
	}

	return psi.GetNodeSnapshot(node).Link, nil
}

func (g *IndexedGraph) OnNodeInvalidated(n psi.Node) {
	g.nodeUpdateQueue <- nodeUpdateRequest{
		Node:    n,
		Version: n.PsiNodeVersion(),
	}
}

func (g *IndexedGraph) OnNodeUpdated(n psi.Node) {
	g.nodeUpdateQueue <- nodeUpdateRequest{
		Node:    n,
		Version: n.PsiNodeVersion(),
	}
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
		close(s.queue)

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

func (g *IndexedGraph) run(proc goprocess.Process) {
	ctx := goprocessctx.OnClosingContext(proc)

	for item := range g.nodeUpdateQueue {
		err := (func(ctx context.Context) error {
			defer func() {
				if r := recover(); r != nil {
					g.logger.Error(r)
				}
			}()

			fn, err := g.store.UpsertNode(ctx, item.Node)

			if err != nil {
				return err
			}

			g.logger.Infow("Updated node", "uuid", item.Node.UUID(), "version", item.Version, "type", typesystem.TypeOf(item.Node).Name(), "cid", fn.Cid)

			return nil
		})(ctx)

		if err != nil {
			g.logger.Error(err)
		} else {
			g.dispatchListeners(item)
		}
	}

	if err := proc.CloseAfterChildren(); err != nil {
		panic(err)
	}
}

func (g *IndexedGraph) dispatchListeners(item nodeUpdateRequest) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, l := range g.listeners {
		l.queue <- item.Node
	}
}
