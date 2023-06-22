package indexing

import (
	"sync"

	"github.com/samber/lo"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type IndexedGraph struct {
	psi.BaseGraph

	mu sync.RWMutex

	nodeMap map[psi.NodeID]psi.Node
}

func NewIndexedGraph() *IndexedGraph {
	g := &IndexedGraph{
		nodeMap: make(map[psi.NodeID]psi.Node),
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

func (g *IndexedGraph) GetNode(id psi.NodeID) (psi.Node, error) {
	if n, ok := g.nodeMap[id]; ok {
		return n, nil
	}

	return nil, psi.ErrNodeNotFound
}

func (g *IndexedGraph) GetNodeChildren(id psi.NodeID) ([]psi.NodeID, error) {
	n, err := g.GetNode(id)

	if err != nil {
		return nil, err
	}

	return lo.Map(n.Children(), func(c psi.Node, _ int) psi.NodeID {
		return c.UUID()
	}), nil
}
