package online

import (
	"context"
	"sync"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/graphfs"
)

var logger = logging.GetLogger("livegraph")

type LiveGraph struct {
	vg *graphfs.VirtualGraph

	mu sync.RWMutex

	nodeCache map[int64]*LiveNode
	pathCache map[string]*LiveNode
}

func NewLiveGraph(vg *graphfs.VirtualGraph) *LiveGraph {
	return &LiveGraph{
		vg: vg,

		nodeCache: map[int64]*LiveNode{},
		pathCache: map[string]*LiveNode{},
	}
}

func (g *LiveGraph) Add(ctx context.Context, node psi.Node) (ln *LiveNode, err error) {
	ln = g.nodeForNode(node)

	return
}

func (g *LiveGraph) Remove(ctx context.Context, n psi.Node) error {
	// TODO: Implement this
}

func (g *LiveGraph) CommitNode(ctx context.Context, node psi.Node) error {
	snap, err := g.Add(ctx, node)

	if err != nil {
		return err
	}

	return snap.Save(ctx)
}

func (g *LiveGraph) ResolveNode(ctx context.Context, path psi.Path) (psi.Node, error) {
	ce, err := g.vg.Resolve(ctx, path)

	if err != nil {
		return nil, err
	}

	if ce.IsNegative() {
		return nil, psi.ErrNodeNotFound
	}

	if ln := g.nodeCache[ce.Inode().ID()]; ln != nil {
		return ln.Get(ctx)
	}

	ln := g.nodeForDentry(ce)

	return ln.Get(ctx)
}

func (g *LiveGraph) nodeForNode(node psi.Node) *LiveNode {
	ln, ok := node.PsiNodeBase().GetSnapshot().(*LiveNode)

	if !ok {
		ln = NewLiveNode(g)

		ln.updateNode(node)
	}

	return ln
}

func (g *LiveGraph) nodeForDentry(ce *graphfs.CacheEntry) *LiveNode {
	nid := ce.Path().String()

	if n := g.pathCache[nid]; n != nil {
		return n
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if n := g.pathCache[nid]; n != nil {
		return n
	}

	ln := NewLiveNode(g)

	ln.updateDentry(ce)

	g.pathCache[nid] = ln

	return ln
}

func (g *LiveGraph) updateNodeCache(ln *LiveNode) {
	ino := int64(-1)
	path := (*psi.Path)(nil)

	if ln.inode != nil {
		ino = ln.inode.ID()
	}

	if ln.dentry != nil {
		p := ln.dentry.Path()
		path = &p
	}

	if ln.cachedIndex != ino {
		if g.nodeCache[ln.cachedIndex] == ln {
			delete(g.nodeCache, ln.cachedIndex)
		}

		ln.cachedIndex = ino

		if ln.cachedIndex >= 0 {
			g.nodeCache[ln.cachedIndex] = ln
		}
	}

	if ln.cachedPath == nil || path == nil || ln.cachedPath.String() != path.String() {
		if ln.cachedPath != nil && g.pathCache[ln.cachedPath.String()] == ln {
			delete(g.pathCache, ln.cachedPath.String())
		}

		ln.cachedPath = path

		if ln.cachedPath != nil {
			g.pathCache[ln.cachedPath.String()] = ln
		}
	}
}
