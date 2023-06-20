package indexing

import "github.com/greenboxal/agibootstrap/pkg/psi"

type IndexedGraph struct {
	psi.BaseGraph
}

func NewIndexedGraph() *IndexedGraph {
	g := &IndexedGraph{}

	g.Init(g)

	return g
}

func (g *IndexedGraph) Add(n psi.Node) {
	g.BaseGraph.Add(n)
}

func (g *IndexedGraph) Remove(n psi.Node) {
	g.BaseGraph.Remove(n)
}
