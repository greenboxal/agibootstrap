package psi

import "gonum.org/v1/gonum/graph/multi"

type Graph struct {
	*multi.DirectedGraph
}

func NewGraph() *Graph {
	return &Graph{
		DirectedGraph: multi.NewDirectedGraph(),
	}
}

func (g *Graph) AddNode(n Node) {
	n.attachToGraph(g)
	g.AddNode(n)
}

func (g *Graph) RemoveNode(n Node) {
	g.RemoveNode(n)
	n.detachFromGraph(nil)
}
