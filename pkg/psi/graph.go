package psi

import (
	"context"
	"sync/atomic"

	"github.com/ipld/go-ipld-prime"
	"gonum.org/v1/gonum/graph/multi"
)

type Graph interface {
	Root() UniqueNode

	Add(n Node)
	Remove(n Node)

	NextNodeID() int64
	NextEdgeID() EdgeID

	SetEdge(e Edge)
	UnsetEdge(self Edge)

	OnNodeUpdated(n Node)
	OnNodeInvalidated(n Node)

	RefreshNode(ctx context.Context, n Node) error
	LoadNode(ctx context.Context, fn *FrozenNode) (Node, error)
	CommitNode(ctx context.Context, node Node) (ipld.Link, error)

	ResolveNode(ctx context.Context, path Path) (n Node, err error)
	ListNodeEdges(ctx context.Context, path Path) (result []*FrozenEdge, err error)
}

type BaseGraph struct {
	self Graph

	g         *multi.DirectedGraph
	nodeIdMap map[NodeID]int64

	nodeIdCounter atomic.Int64
	edgeIdCounter atomic.Int64
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
	n.PsiNodeBase().AttachToGraph(g.self)
	g.g.AddNode(n)
	g.nodeIdMap[n.CanonicalPath().String()] = n.ID()
}

func (g *BaseGraph) Remove(n Node) {
	g.g.RemoveNode(n.ID())
	n.PsiNodeBase().DetachFromGraph(g.self)
	delete(g.nodeIdMap, n.CanonicalPath().String())
}

func (g *BaseGraph) SetEdge(e Edge) {
	g.g.SetLine(psiEdgeWrapper{e})
}

func (g *BaseGraph) UnsetEdge(self Edge) {
	g.g.RemoveLine(self.From().ID(), self.To().ID(), int64(self.ID()))
}

func (g *BaseGraph) OnNodeInvalidated(n Node) {}
func (g *BaseGraph) OnNodeUpdated(n Node)     {}
