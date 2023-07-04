package psi

import (
	"fmt"

	"gonum.org/v1/gonum/graph"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
)

type EdgeID int64

type Edge interface {
	ID() EdgeID
	Key() EdgeReference
	Kind() EdgeKind
	From() Node
	To() Node

	ReplaceTo(node Node) Edge
	ReplaceFrom(node Node) Edge

	attachToGraph(g Graph)
	detachFromGraph(g Graph)
}

type EdgeBase struct {
	g Graph

	id   EdgeID
	key  EdgeReference
	from Node
	to   Node
}

func NewEdgeBase(key EdgeReference, from Node, to Node) *EdgeBase {
	return &EdgeBase{
		key:  key,
		from: from,
		to:   to,
	}
}

func (e *EdgeBase) ID() EdgeID         { return e.id }
func (e *EdgeBase) Key() EdgeReference { return e.key }
func (e *EdgeBase) Kind() EdgeKind     { return e.key.GetKind() }
func (e *EdgeBase) From() Node         { return e.from }
func (e *EdgeBase) To() Node           { return e.to }

func (e *EdgeBase) String() string {
	return fmt.Sprintf("Edge(%s): %s -> %s", e.key, e.from, e.to)
}

func (e *EdgeBase) ReplaceTo(node Node) Edge {
	return &EdgeBase{
		key:  e.key,
		from: e.from,
		to:   node,
	}
}
func (e *EdgeBase) ReplaceFrom(node Node) Edge {
	return &EdgeBase{
		key:  e.key,
		from: node,
		to:   e.to,
	}
}

func (e *EdgeBase) attachToGraph(g Graph) {
	if e.g == g {
		return
	}

	if g == nil {
		e.detachFromGraph(nil)
		return
	}

	if e.g != nil {
		panic("node already attached to a graph")
	}

	e.g = g

	e.from.attachToGraph(g)
	e.to.attachToGraph(g)

	if e.g != nil {
		e.id = g.NextEdgeID()

		e.g.SetEdge(e)
	}
}

func (e *EdgeBase) detachFromGraph(g Graph) {
	if e.g == nil {
		return
	}

	if e.g != g {
		return
	}

	oldGraph := e.g

	e.g = nil

	oldGraph.UnsetEdge(e)
}

type EdgeIterator interface {
	Next() bool
	Value() Edge
	Edge() Edge
}

type nodeEdgeIterator struct {
	n       *NodeBase
	it      collectionsfx.MapIterator[EdgeKey, Edge]
	current Edge
}

func (e *nodeEdgeIterator) Value() Edge { return e.Edge() }
func (e *nodeEdgeIterator) Edge() Edge  { return e.current }

func (e *nodeEdgeIterator) Next() bool {
	if e.it == nil {
		e.it = e.n.edges.MapIterator()
	}

	if !e.it.Next() {
		return false
	}

	e.current = e.it.Value()

	return true
}

type psiEdgeWrapper struct{ Edge }

func (p psiEdgeWrapper) From() graph.Node {
	return p.Edge.From()
}

func (p psiEdgeWrapper) To() graph.Node {
	return p.Edge.To()
}

func (p psiEdgeWrapper) ID() int64 {
	return int64(p.Edge.ID())
}

func (p psiEdgeWrapper) ReversedLine() graph.Line {
	panic("not supported")
}
