package psi

import (
	"context"
	"fmt"
	"sync"

	"github.com/ipld/go-ipld-prime"
)

type EdgeID int64

type Edge interface {
	ID() EdgeID
	Key() EdgeReference
	Kind() EdgeKind

	PsiEdgeBase() *EdgeBase

	From() Node
	To() Node
	ResolveTo(ctx context.Context) (Node, error)

	ReplaceTo(node Node) Edge

	attachToGraph(g Graph)
	detachFromGraph(g Graph)
}

type EdgeSnapshot struct {
	Frozen *FrozenEdge
	Link   ipld.Link
	Node   Node
	Index  int64
}

type EdgeBase struct {
	g    Graph
	self Edge
	snap EdgeSnapshot

	id EdgeID
}

func (e *EdgeBase) ID() EdgeID             { return e.id }
func (e *EdgeBase) Kind() EdgeKind         { return e.self.Key().GetKind() }
func (e *EdgeBase) Graph() Graph           { return e.g }
func (e *EdgeBase) PsiEdgeBase() *EdgeBase { return e }

func (e *EdgeBase) String() string {
	return fmt.Sprintf("Edge(%d, %s)", e.id, e.self.Key())
}

func (e *EdgeBase) Init(self Edge) {
	if e.self != nil {
		panic("edge already initialized")
	}

	e.self = self
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
}

func (e *EdgeBase) detachFromGraph(g Graph) {
	if e.g == nil {
		return
	}

	if e.g != g {
		return
	}

	e.g = nil
}

func (e *EdgeBase) GetSnapshot() EdgeSnapshot     { return e.snap }
func (e *EdgeBase) SetSnapshot(snap EdgeSnapshot) { e.snap = snap }

type LazyEdge struct {
	EdgeBase

	mu   sync.RWMutex
	cond *sync.Cond

	key  EdgeReference
	from Node
	to   Node

	valid    bool
	resolver ResolveEdgeFunc
}

func NewLazyEdge(g Graph, key EdgeReference, from Node, resolver ResolveEdgeFunc) Edge {
	le := &LazyEdge{}
	le.g = g
	le.cond = sync.NewCond(le.mu.RLocker())
	le.key = key
	le.from = from
	le.resolver = resolver
	le.Init(le)
	return le
}

func (l *LazyEdge) Key() EdgeReference { return l.key }
func (l *LazyEdge) From() Node         { return l.from }

func (l *LazyEdge) To() Node {
	n, err := l.ResolveTo(context.Background())

	if err != nil {
		panic(err)
	}

	return n
}

func (l *LazyEdge) ResolveTo(ctx context.Context) (Node, error) {
	if l.valid {
		return l.to, nil
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	if !l.valid {
		n, err := l.resolver(ctx, l.g, l.from, l.key.GetKey())

		if err != nil {
			return nil, err
		}

		l.to = n
		l.valid = true
	}

	return l.to, nil
}

func (l *LazyEdge) ReplaceTo(node Node) Edge {
	return NewSimpleEdge(l.key, l.from, node)
}

func (l *LazyEdge) Invalidate() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.valid {
		l.valid = false
	}
}

type SimpleEdge struct {
	EdgeBase

	key  EdgeReference
	from *NodeBase
	to   *NodeBase
}

func NewSimpleEdge(key EdgeReference, from Node, to Node) Edge {
	se := &SimpleEdge{}
	se.key = key
	se.from = from.PsiNodeBase()
	se.to = to.PsiNodeBase()
	se.Init(se)
	return se
}

func (e *SimpleEdge) Key() EdgeReference { return e.key }
func (e *SimpleEdge) From() Node         { return e.from.PsiNode() }
func (e *SimpleEdge) To() Node           { return e.to.PsiNode() }

func (e *SimpleEdge) ResolveTo(ctx context.Context) (Node, error) { return e.To(), nil }

func (e *SimpleEdge) ReplaceTo(node Node) Edge {
	return NewSimpleEdge(e.key, e.from, node)
}
