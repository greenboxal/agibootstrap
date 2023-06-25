package psi

import "fmt"

type EdgeID int64

type EdgeKind string

func (k EdgeKind) String() string {
	return string(k)
}

type TypedEdgeKind[T Node] EdgeKind

func (f TypedEdgeKind[T]) Singleton() TypedEdgeKey[T] {
	return TypedEdgeKey[T]{Kind: f}
}

type EdgeReference interface {
	GetKind() EdgeKind
	GetName() string
	GetIndex() int64
	GetKey() EdgeKey
}

type EdgeKey struct {
	Kind  EdgeKind `json:"Kind"`
	Name  string   `json:"Name"`
	Index int64    `json:"Index"`
}

func (k EdgeKey) GetKind() EdgeKind { return k.Kind }
func (k EdgeKey) GetName() string   { return k.Name }
func (k EdgeKey) GetIndex() int64   { return k.Index }
func (k EdgeKey) GetKey() EdgeKey   { return k }
func (k EdgeKey) String() string    { return fmt.Sprintf("%s=%d:%s", k.Kind, k.Index, k.Name) }

type TypedEdgeKey[T Node] struct {
	Kind  TypedEdgeKind[T] `json:"Kind"`
	Name  string           `json:"Name"`
	Index int64            `json:"Index"`
}

func (k TypedEdgeKey[T]) GetKind() EdgeKind { return EdgeKind(k.Kind) }
func (k TypedEdgeKey[T]) GetName() string   { return k.Name }
func (k TypedEdgeKey[T]) GetIndex() int64   { return k.Index }
func (k TypedEdgeKey[T]) String() string    { return fmt.Sprintf("%s=%d:%s", k.Kind, k.Index, k.Name) }

func (k TypedEdgeKey[T]) GetKey() EdgeKey {
	return EdgeKey{
		Kind:  k.GetKind(),
		Name:  k.GetName(),
		Index: k.GetIndex(),
	}
}

type Edge interface {
	ID() EdgeID
	Key() EdgeReference
	Kind() EdgeKind
	From() Node
	To() Node

	ReplaceTo(node Node) Edge
	ReplaceFrom(node Node) Edge
	attachToGraph(g Graph)
}

type EdgeBase struct {
	id   EdgeID
	key  EdgeReference
	from Node
	to   Node
}

func (e *EdgeBase) String() string {
	return fmt.Sprintf("Edge(%s): %s -> %s", e.key, e.from, e.to)
}

func (e *EdgeBase) ID() EdgeID         { return e.id }
func (e *EdgeBase) Key() EdgeReference { return e.key }
func (e *EdgeBase) Kind() EdgeKind     { return e.key.GetKind() }
func (e *EdgeBase) From() Node         { return e.from }
func (e *EdgeBase) To() Node           { return e.to }

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
	e.from.attachToGraph(g)
	e.to.attachToGraph(g)
}

type EdgeIterator interface {
	Next() bool
	Edge() Edge
}

type edgeIterator struct {
	n       *NodeBase
	keys    []EdgeKey
	index   int
	current Edge
}

func (e edgeIterator) Next() bool {
	if e.index >= len(e.keys) {
		return false
	}

	k := e.keys[e.index]
	e.current = e.n.edges[k]

	e.index++

	return true
}

func (e edgeIterator) Edge() Edge {
	return e.current
}
