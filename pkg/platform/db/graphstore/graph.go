package graphstore

import (
	"context"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"github.com/samber/lo"
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

type cachedNode struct {
	mu sync.Mutex

	uuid   psi.NodeID
	frozen *psi.FrozenNode
	node   psi.Node
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

	root psi.Node

	ds    datastore.Batching
	store *Store
	wal   *WriteAheadLog

	nodeCache map[psi.NodeID]*cachedNode

	proc            goprocess.Process
	nodeUpdateQueue chan nodeUpdateRequest

	listeners []*listenerSlot
}

func NewIndexedGraph(ds datastore.Batching, walPath string, root psi.Node) (*IndexedGraph, error) {
	if err := os.MkdirAll(walPath, 0755); err != nil {
		return nil, err
	}

	wal, err := NewWriteAheadLog(walPath)

	if err != nil {
		return nil, err
	}

	store := NewStore(ds, wal)

	g := &IndexedGraph{
		logger: logging.GetLogger("graphstore"),

		ds:    ds,
		wal:   wal,
		store: store,
		root:  root,

		nodeCache: map[psi.NodeID]*cachedNode{},

		nodeUpdateQueue: make(chan nodeUpdateRequest, 256),
	}

	g.Init(g)

	g.proc = goprocess.Go(g.run)

	return g, nil
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

		if _, err := g.CommitNode(context.Background(), n); err != nil {
			panic(err)
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

func (g *IndexedGraph) LoadNode(ctx context.Context, fn *psi.FrozenNode) (psi.Node, error) {
	entry := g.getCacheEntry(fn.UUID, true)

	typ := psi.NodeTypeByName(fn.Type)

	if typ.Definition().IsRuntimeOnly {
		if entry == nil {
			return nil, psi.ErrNodeNotFound
		}

		return entry.node, nil
	}

	rawNode, err := g.store.lsys.Load(ipld.LinkContext{Ctx: ctx}, fn.Link, typ.Type().IpldPrototype())

	if err != nil {
		return nil, err
	}

	n := typesystem.Unwrap(rawNode).(psi.Node)

	for k, v := range fn.Attributes {
		n.SetAttribute(k, v)
	}

	for _, edgeLink := range fn.Edges {
		rawEdge, err := g.store.lsys.Load(ipld.LinkContext{Ctx: ctx}, edgeLink, frozenEdgeType.IpldPrototype())

		if err != nil {
			return nil, err
		}

		fe := typesystem.Unwrap(rawEdge).(psi.FrozenEdge)

		to, err := g.GetNodeByID(fe.To)

		if err != nil {
			return nil, err
		}

		if fe.Key.Kind == psi.EdgeKindChild {
			idx := fe.Key.Index

			if idx >= int64(len(n.Children())) {
				idx = int64(len(n.Children()))
			}

			n.InsertChildrenAt(int(fe.Key.Index), to)
		} else {
			n.SetEdge(fe.Key, to)
		}
	}

	return n, err
}

func (g *IndexedGraph) CommitNode(ctx context.Context, node psi.Node) (ipld.Link, error) {
	fn, edges, link, err := g.store.FreezeNode(ctx, node)

	if err != nil {
		panic(err)
	}

	records := make([]WalRecord, 2+len(edges))

	records[0] = BuildWalRecord(WalOpUpdateNode, link.(cidlink.Link).Cid)

	for i, edgeLink := range fn.Edges {
		records[i+1] = BuildWalRecord(WalOpUpdateEdge, edgeLink.Cid)
	}

	records[len(records)-1] = BuildWalRecord(WalOpFence, cid.Undef)

	rid, err := g.wal.WriteRecords(records...)

	if err != nil {
		return nil, err
	}

	psi.UpdateNodeSnapshot(node, &psi.NodeSnapshot{
		Timestamp: time.Now(),
		Fence:     rid,
		Link:      link,
	})

	g.nodeUpdateQueue <- nodeUpdateRequest{
		Fence:  rid,
		Node:   node,
		Frozen: fn,
		Edges:  edges,
		Link:   link,
	}

	return link, nil
}

func (g *IndexedGraph) OnNodeInvalidated(n psi.Node) {
}

func (g *IndexedGraph) OnNodeUpdated(n psi.Node) {
	if _, err := g.CommitNode(context.Background(), n); err != nil {
		panic(err)
	}
}

func (g *IndexedGraph) indexNode(ctx context.Context, item nodeUpdateRequest) error {
	batch, err := g.store.ds.Batch(ctx)

	if err != nil {
		return err
	}

	if err := g.store.batchUpsertNode(ctx, batch, item.Frozen, item.Link); err != nil {
		return err
	}

	for i, edge := range item.Edges {
		link := item.Frozen.Edges[i]

		if err := g.store.batchUpsertEdge(ctx, batch, edge, link); err != nil {
			return err
		}
	}

	if err := batch.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (g *IndexedGraph) run(proc goprocess.Process) {
	ctx := goprocessctx.OnClosingContext(proc)

	lastFenceData, err := g.ds.Get(ctx, lastFenceKey)

	if err == nil && len(lastFenceData) > 0 {
		lastFence, err := strconv.ParseUint(string(lastFenceData), 10, 64)

		if err != nil {
			panic(err)
		}

		for i := lastFence; i <= g.wal.LastRecordIndex(); i++ {
			rec, err := g.wal.ReadRecord(i)

			if err != nil {
				panic(err)
			}

			if rec.Op != WalOpUpdateNode {
				continue
			}

			fnLink := cidlink.Link{Cid: rec.Payload}
			fn, err := g.store.GetNodeByCid(ctx, fnLink)

			if err != nil {
				panic(err)
			}

			edges := make([]*psi.FrozenEdge, len(fn.Edges))

			for i, edgeLink := range fn.Edges {
				edge, err := g.store.GetEdgeByCid(ctx, edgeLink)

				if err != nil {
					panic(err)
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
	} else if err != nil && err != datastore.ErrNotFound {
		panic(err)
	}

	for {
		select {
		case <-ctx.Done():
			if err := proc.CloseAfterChildren(); err != nil {
				panic(err)
			}

			return

		case item := <-g.nodeUpdateQueue:
			if err := g.processQueueItem(ctx, item); err != nil {
				g.logger.Error(err)
			}
		}
	}
}

func (g *IndexedGraph) processQueueItem(ctx context.Context, item nodeUpdateRequest) error {
	if item.Node == nil {
		n, err := g.LoadNode(ctx, item.Frozen)

		if err != nil {
			return err
		}

		item.Node = n
	}

	if err := g.indexNode(ctx, item); err != nil {
		return err
	}

	err := g.ds.Put(ctx, lastFenceKey, []byte(strconv.FormatUint(item.Fence, 10)))

	if err != nil {
		return err
	}

	g.dispatchListeners(item)

	return nil
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
		node, err := g.LoadNode(ctx, entry.frozen)

		if err != nil {
			return err
		}

		entry.node = node
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
		defer close(s.queue)

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

func (g *IndexedGraph) dispatchListeners(item nodeUpdateRequest) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, l := range g.listeners {
		l.queue <- item.Node
	}
}

func (g *IndexedGraph) Close() error {
	if g.proc != nil {
		if err := g.proc.Close(); err != nil {
			return err
		}

		g.proc = nil
	}

	if g.wal != nil {
		if err := g.wal.Close(); err != nil {
			return err
		}

		g.wal = nil
	}

	return nil
}
