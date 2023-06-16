package psi

import (
	"gonum.org/v1/gonum/graph/multi"
)

type Graph struct {
	g *multi.DirectedGraph
}

func NewGraph() *Graph {
	return &Graph{
		g: multi.NewDirectedGraph(),
	}
}

func (g *Graph) Add(n Node) {
	n.attachToGraph(g)
	g.g.AddNode(n)
}

func (g *Graph) Remove(n Node) {
	g.g.RemoveNode(n.ID())
	n.detachFromGraph(nil)
}

func (g *Graph) Replace(old, new Node) {
	if old == new {
		return
	}

	gn := old.Node().g

	if gn != nil && gn != g {
		panic("nodes belong to different graphs")
	}

	parent := old.Parent()

	if parent != nil {
		parent.replaceChildNode(old, new)
	}
}
