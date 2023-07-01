package psi

import (
	"reflect"
	"sync/atomic"

	"gonum.org/v1/gonum/graph/multi"
)

type Graph interface {
	Add(n Node)
	Remove(n Node)
	Replace(old, new Node)

	NextNodeID() int64
	NextEdgeID() EdgeID

	Nodes() NodeIterator
	Edges() EdgeIterator

	SetEdge(e Edge)
	UnsetEdge(self Edge)

	OnNodeUpdated(n Node)
	OnNodeInvalidated(n Node)
}

type BaseGraph struct {
	self Graph

	g         *multi.DirectedGraph
	nodeIdMap map[NodeID]int64

	nodeIdCounter atomic.Int64
	edgeIdCounter atomic.Int64
}

func (g *BaseGraph) Nodes() NodeIterator {
	return &graphNodeIterator{g: g}
}

func (g *BaseGraph) Edges() EdgeIterator {
	return &graphEdgeIterator{g: g}
}

func (g *BaseGraph) Init(self Graph) {
	g.self = self

	g.g = multi.NewDirectedGraph()
	g.g.AddNode(g.g.NewNode())

	g.nodeIdMap = make(map[NodeID]int64)
}

func (g *BaseGraph) NextNodeID() int64 {
	return g.nodeIdCounter.Add(1)
}

func (g *BaseGraph) NextEdgeID() EdgeID {
	return EdgeID(g.edgeIdCounter.Add(1))
}

func (g *BaseGraph) Add(n Node) {
	n.attachToGraph(g.self)
	g.g.AddNode(n)
	g.nodeIdMap[n.UUID()] = n.ID()
}

func (g *BaseGraph) Remove(n Node) {
	g.g.RemoveNode(n.ID())
	n.detachFromGraph(nil)
	delete(g.nodeIdMap, n.UUID())
}

func (g *BaseGraph) SetEdge(e Edge) {
	g.g.SetLine(psiEdgeWrapper{e})
}

func (g *BaseGraph) UnsetEdge(self Edge) {
	g.g.RemoveLine(self.From().ID(), self.To().ID(), int64(self.ID()))
}

func (g *BaseGraph) Replace(old, new Node) {
	if old == new {
		return
	}

	gn := old.PsiNodeBase().g

	if gn != nil && gn != g {
		panic("nodes belong to different graphs")
	}

	parent := old.Parent()

	if parent != nil {
		parent.ReplaceChildNode(old, new)
	}
}

func (g *BaseGraph) OnNodeInvalidated(n Node) {}
func (g *BaseGraph) OnNodeUpdated(n Node)     {}

type graphNodeIterator struct {
	g       *BaseGraph
	current Node
	iter    *reflect.MapIter
}

func (g *graphNodeIterator) Value() Node { return g.current }

func (g *graphNodeIterator) Next() bool {
	if g.iter == nil {
		g.iter = reflect.ValueOf(g.g.nodeIdMap).MapRange()
	}

	for {
		if !g.iter.Next() {
			return false
		}

		id := g.iter.Value().Int()
		n, ok := g.g.g.Node(id).(Node)

		if !ok {
			continue
		}

		g.current = n

		return true
	}
}

type graphEdgeIterator struct {
	g       *BaseGraph
	nodes   NodeIterator
	edges   EdgeIterator
	current Edge
}

func (g *graphEdgeIterator) Value() Edge { return g.Edge() }
func (g *graphEdgeIterator) Edge() Edge  { return g.current }

func (g *graphEdgeIterator) Next() bool {
	for g.edges == nil || !g.edges.Next() {
		if g.nodes == nil {
			g.nodes = g.g.Nodes()
		}

		if !g.nodes.Next() {
			return false
		}

		g.edges = g.nodes.Value().Edges()
	}

	g.current = g.edges.Edge()

	return true
}
