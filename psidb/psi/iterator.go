package psi

import (
	"gonum.org/v1/gonum/graph"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
)

type NodeIterator interface {
	iterators.Iterator[Node]

	Value() Node
	Next() bool
}

type EdgeIterator interface {
	iterators.Iterator[Edge]

	Next() bool
	Value() Edge
}

type nodeSliceIterator struct {
	current Node
	items   []Node
}

func (n *nodeSliceIterator) Value() Node { return n.current }

func (n *nodeSliceIterator) Next() bool {
	if len(n.items) == 0 {
		return false
	}

	n.current = n.items[0]
	n.items = n.items[1:]

	return true
}

func (n *nodeSliceIterator) Prepend(iterator NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: []NodeIterator{iterator, n}}
}

func (n *nodeSliceIterator) Append(iterator NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: []NodeIterator{n, iterator}}
}

type nodeChildrenIterator struct {
	parent  *NodeBase
	current Node
	index   int
}

func (n *nodeChildrenIterator) Value() Node { return n.current }

func (n *nodeChildrenIterator) Next() bool {
	if n.index >= n.parent.children.Len() {
		return false
	}

	n.current = n.parent.children.Get(n.index)
	n.index++

	return true
}

func (n *nodeChildrenIterator) Prepend(iterator NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: []NodeIterator{iterator, n}}
}

func (n *nodeChildrenIterator) Append(iterator NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: []NodeIterator{n, iterator}}
}

type nestedNodeIterator struct {
	current   NodeIterator
	iterators []NodeIterator
}

func (n *nestedNodeIterator) Value() Node {
	if n.current == nil {
		return nil
	}

	return n.current.Value()
}

func (n *nestedNodeIterator) Next() bool {
	for n.current == nil || !n.current.Next() {
		if len(n.iterators) == 0 {
			return false
		}

		n.current = n.iterators[0]
		n.iterators = n.iterators[1:]
	}

	return true
}

func (n *nestedNodeIterator) Prepend(iterator NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: []NodeIterator{iterator, n}}
}

func (n *nestedNodeIterator) Append(iterator NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: []NodeIterator{n, iterator}}
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
