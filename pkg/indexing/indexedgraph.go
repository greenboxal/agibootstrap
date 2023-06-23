package indexing

import (
	"sync"

	"github.com/samber/lo"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type IndexedGraph struct {
	psi.BaseGraph

	mu   sync.RWMutex
	root psi.Node

	nodeMap map[psi.NodeID]psi.Node
	pathMap map[string]psi.Node
}

func NewIndexedGraph(root psi.Node) *IndexedGraph {
	g := &IndexedGraph{
		root:    root,
		nodeMap: make(map[psi.NodeID]psi.Node),
		pathMap: make(map[string]psi.Node),
	}

	g.Init(g)

	return g
}

func (g *IndexedGraph) Add(n psi.Node) {
	if _, ok := g.nodeMap[n.UUID()]; ok {
		return
	}

	g.nodeMap[n.UUID()] = n

	g.BaseGraph.Add(n)
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
