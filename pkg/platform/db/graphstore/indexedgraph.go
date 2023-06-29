package graphstore

import (
	"context"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type nodeUpdateRequest struct {
	Node    psi.Node
	Version int64
}

type IndexedGraph struct {
	psi.BaseGraph

	logger *zap.SugaredLogger
	mu     sync.RWMutex

	store *Store
	root  psi.Node

	nodeMap map[psi.NodeID]psi.Node
	pathMap map[string]psi.Node

	proc            goprocess.Process
	nodeUpdateQueue chan nodeUpdateRequest
}

func NewIndexedGraph(ctx context.Context, ds datastore.Batching, root psi.Node) *IndexedGraph {
	os := NewObjectStore(ds)
	store := NewStore(ds, os)

	g := &IndexedGraph{
		logger: logging.GetLogger("graphstore"),

		root:  root,
		store: store,

		nodeMap: make(map[psi.NodeID]psi.Node),
		pathMap: make(map[string]psi.Node),

		nodeUpdateQueue: make(chan nodeUpdateRequest, 100),
	}

	g.Init(g)

	g.proc = goprocess.Go(g.run)

	return g
}

func (g *IndexedGraph) Add(n psi.Node) {
	if _, ok := g.nodeMap[n.UUID()]; ok {
		return
	}

	g.nodeMap[n.UUID()] = n

	g.BaseGraph.Add(n)

	if _, err := g.store.UpsertNode(context.Background(), n); err != nil {
		panic(err)
	}
}

func (g *IndexedGraph) Remove(n psi.Node) {
	delete(g.nodeMap, n.UUID())

	g.BaseGraph.Remove(n)
}

func (g *IndexedGraph) ResolveNode(path psi.Path) (n psi.Node, err error) {
	return psi.ResolvePath(g.root, path)
}

func (g *IndexedGraph) GetNodeByID(id psi.NodeID) (psi.Node, error) {
	if n, ok := g.nodeMap[id]; ok {
		return n, nil
	}

	return nil, psi.ErrNodeNotFound
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

func (g *IndexedGraph) run(proc goprocess.Process) {
	ctx := goprocessctx.OnClosingContext(proc)

	for item := range g.nodeUpdateQueue {
		fn, err := g.store.UpsertNode(ctx, item.Node)

		if err != nil {
			g.logger.Error(err)
		}

		g.logger.Infow("Updated node", "uuid", item.Node.UUID(), "version", item.Version, "cid", fn.Cid)
	}
}
