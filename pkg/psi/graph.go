package psi

import (
	"sync/atomic"

	"gonum.org/v1/gonum/graph/multi"
)

type Graph interface {
	Add(n Node)
	Remove(n Node)
	Replace(old, new Node)
	AllocateNodeID() int64
}

type BaseGraph struct {
	g         *multi.DirectedGraph
	self      Graph
	idCounter atomic.Int64
}

func (g *BaseGraph) Init(self Graph) {
	g.self = self
	g.g = multi.NewDirectedGraph()
	g.g.AddNode(g.g.NewNode())
}

func (g *BaseGraph) AllocateNodeID() int64 {
	return g.idCounter.Add(1)
}

func (g *BaseGraph) Add(n Node) {
	n.attachToGraph(g.self)
	g.g.AddNode(n)
}

func (g *BaseGraph) Remove(n Node) {
	g.g.RemoveNode(n.ID())
	n.detachFromGraph(nil)
}

func (g *BaseGraph) Replace(old, new Node) {
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
