package psi

type EdgeID int64

type EdgeKind string

type EdgeKey struct {
	Kind EdgeKind
	Name string
}

type Edge interface {
	ID() EdgeID
	Key() EdgeKey
	Kind() EdgeKind
	From() Node
	To() Node

	ReplaceTo(node Node) Edge
	ReplaceFrom(node Node) Edge
}

type EdgeBase struct {
	id   EdgeID
	key  EdgeKey
	from Node
	to   Node
}

func (e *EdgeBase) ID() EdgeID     { return e.id }
func (e *EdgeBase) Key() EdgeKey   { return e.key }
func (e *EdgeBase) Kind() EdgeKind { return e.key.Kind }
func (e *EdgeBase) From() Node     { return e.from }
func (e *EdgeBase) To() Node       { return e.to }

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
