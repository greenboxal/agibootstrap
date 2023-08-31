package online

import (
	"context"
	"fmt"
	"sync"

	"github.com/ipld/go-ipld-prime/linking"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	`github.com/greenboxal/agibootstrap/psidb/core/api`
	graphfs "github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

var logger = logging.GetLogger("livegraph")
var tracer = otel.Tracer("livegraph", trace.WithInstrumentationAttributes(
	semconv.ServiceName("psidb-graph"),
))

type LiveGraph struct {
	mu sync.RWMutex

	graph *graphfs.VirtualGraph
	lsys  *linking.LinkSystem
	tx    *graphfs.Transaction
	root  psi.Path

	dirtySet  map[psi.Node]struct{}
	nodeCache map[int64]*LiveNode
	pathCache map[string]*LiveNode

	sp           inject.ServiceProvider
	typeRegistry psi.TypeRegistry
}

var _ psi.Graph = (*LiveGraph)(nil)

func NewLiveGraph(
	ctx context.Context,
	root psi.Path,
	vg *graphfs.VirtualGraph,
	lsys *linking.LinkSystem,
	types psi.TypeRegistry,
	sp inject.ServiceLocator,
) (*LiveGraph, error) {
	tx, err := vg.BeginTransaction(ctx)

	if err != nil {
		return nil, err
	}

	lg := &LiveGraph{
		graph:        vg,
		root:         root,
		lsys:         lsys,
		tx:           tx,
		typeRegistry: types,

		nodeCache: map[int64]*LiveNode{},
		pathCache: map[string]*LiveNode{},
		dirtySet:  map[psi.Node]struct{}{},
	}

	lg.sp = inject.NewServiceProvider(inject.WithParentServiceProvider(sp))

	inject.RegisterInstance[psi.TypeRegistry](lg.sp, lg.typeRegistry)
	inject.RegisterInstance[psi.Graph](lg.sp, lg)

	return lg, nil
}

func (lg *LiveGraph) Root() psi.Path                          { return lg.root }
func (lg *LiveGraph) ServiceProvider() inject.ServiceProvider { return lg.sp }
func (lg *LiveGraph) Services() inject.ServiceLocator         { return lg.sp }
func (lg *LiveGraph) Transaction() *graphfs.Transaction       { return lg.tx }

func (lg *LiveGraph) Add(node psi.Node) {
	_, err := lg.addLiveNode(node)

	if err != nil {
		panic(err)
	}
}

func (lg *LiveGraph) Remove(node psi.Node) {
	if err := lg.Delete(context.Background(), node); err != nil {
		panic(err)
	}
}

func (lg *LiveGraph) markDirty(node psi.Node) {
	ln := lg.nodeForNode(node)

	ln.flags |= liveNodeFlagDirty

	lg.dirtySet[node] = struct{}{}
}

func (lg *LiveGraph) markClean(node psi.Node) {
	ln := lg.nodeForNode(node)

	delete(lg.dirtySet, node)

	ln.flags &= ^liveNodeFlagDirty
}

func (lg *LiveGraph) addLiveNode(node psi.Node) (ln *LiveNode, err error) {
	if node.PsiNode() == nil {
		return nil, fmt.Errorf("node is not initialized")
	}

	if g := node.PsiNodeBase().Graph(); g != nil && g != lg {
		return nil, fmt.Errorf("node is already attached to a different graph")
	}

	ln = lg.nodeForNode(node)

	node.PsiNodeBase().AttachToGraph(lg)

	return
}

func (lg *LiveGraph) Delete(ctx context.Context, n psi.Node) error {
	// TODO: Implement this
	return nil
}

func (lg *LiveGraph) Resolve(ctx context.Context, path psi.Path) (psi.Node, error) {
	return lg.ResolveNode(ctx, path)
}

func (lg *LiveGraph) ResolveNode(ctx context.Context, path psi.Path) (psi.Node, error) {
	ln, err := lg.resolveNodeUnloaded(ctx, path)

	if err != nil {
		return nil, err
	}

	return ln.Get(ctx)
}

func (lg *LiveGraph) resolveNodeUnloaded(ctx context.Context, path psi.Path) (*LiveNode, error) {
	if path.IsRelative() {
		path = lg.root.Join(path)
	}

	if path.IsRelative() {
		return nil, fmt.Errorf("path must be absolute")
	}

	ce, err := lg.graph.Resolve(ctx, path)

	if err != nil {
		return nil, err
	}

	if ce.IsNegative() {
		return nil, psi.ErrNodeNotFound
	}

	if ln := lg.nodeCache[ce.Inode().ID()]; ln != nil {
		return ln, nil
	}

	ln := lg.nodeForDentry(ce)

	return ln, nil
}

func (lg *LiveGraph) ListNodeEdges(ctx context.Context, path psi.Path) (result []*psi.FrozenEdge, err error) {
	if path.IsRelative() {
		return nil, fmt.Errorf("path must be absolute")
	}

	edges, err := lg.graph.ReadEdges(ctx, path)

	if err != nil {
		return nil, err
	}

	result = iterators.ToSlice(iterators.Map(edges, func(edge *coreapi.SerializedEdge) *psi.FrozenEdge {
		return &psi.FrozenEdge{
			Key:     edge.Key,
			ToPath:  &edge.ToPath,
			ToIndex: edge.ToIndex,
		}
	}))

	return result, nil
}

func (lg *LiveGraph) Commit(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "LiveGraph.Commit")
	defer span.End()

	for ln, _ := range lg.dirtySet {
		if err := ln.Update(ctx); err != nil {
			return err
		}
	}

	return lg.tx.Commit(ctx)
}

func (lg *LiveGraph) Rollback(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "LiveGraph.Rollback")
	defer span.End()

	return lg.tx.Rollback(ctx)
}

func (lg *LiveGraph) nodeForNode(node psi.Node) *LiveNode {
	ln, ok := node.PsiNodeBase().GetSnapshot().(*LiveNode)

	if !ok {
		ln = NewLiveNode(lg)
	}

	if ln.node == nil {
		ln.updateNode(node)
	}

	return ln
}

func (lg *LiveGraph) nodeForDentry(ce *graphfs.CacheEntry) *LiveNode {
	nid := ce.Path().String()

	if n := lg.pathCache[nid]; n != nil {
		return n
	}

	lg.mu.Lock()
	defer lg.mu.Unlock()

	if n := lg.pathCache[nid]; n != nil {
		return n
	}

	ln := NewLiveNode(lg)
	ln.path = ce.Path()

	ln.updateDentry(ce)

	lg.pathCache[nid] = ln

	return ln
}

func (lg *LiveGraph) updateNodeCache(ln *LiveNode) {
	ino := int64(-1)

	if ln.inode != nil {
		ino = ln.inode.ID()
	}

	if ln.cachedIndex != ino {
		if lg.nodeCache[ln.cachedIndex] == ln {
			delete(lg.nodeCache, ln.cachedIndex)
		}

		ln.cachedIndex = ino

		if ln.cachedIndex >= 0 {
			lg.nodeCache[ln.cachedIndex] = ln
		}
	}

	if ln.cachedPath == nil || ln.cachedPath.String() != ln.path.String() {
		if ln.cachedPath != nil && lg.pathCache[ln.cachedPath.String()] == ln {
			delete(lg.pathCache, ln.cachedPath.String())
		}

		ln.cachedPath = &ln.path

		if ln.cachedPath != nil {
			lg.pathCache[ln.cachedPath.String()] = ln
		}
	}
}

func (lg *LiveGraph) Close() error {
	for _, n := range lg.nodeCache {
		if err := n.Close(); err != nil {
			return err
		}
	}

	return nil
}
