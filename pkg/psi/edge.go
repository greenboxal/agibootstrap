package psi

import (
	"fmt"
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

func (e *EdgeBase) SetFrom(from Node)        { e.from = from }
func (e *EdgeBase) SetTo(to Node)            { e.to = to }
func (e *EdgeBase) SetKey(key EdgeReference) { e.key = key.GetKey() }

func (e *EdgeBase) ID() EdgeID         { return e.id }
func (e *EdgeBase) Key() EdgeReference { return e.key }
func (e *EdgeBase) Kind() EdgeKind     { return e.key.GetKind() }
func (e *EdgeBase) From() Node         { return e.from }
func (e *EdgeBase) To() Node           { return e.to }
func (e *EdgeBase) Graph() Graph       { return e.g }

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

	if e.from != nil {
		e.from.attachToGraph(g)
	}

	if e.to != nil {
		e.to.attachToGraph(g)
	}

	if e.g != nil {
		e.id = g.NextEdgeID()

		if e.from != nil && e.to != nil {
			e.g.SetEdge(e)
		}
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
