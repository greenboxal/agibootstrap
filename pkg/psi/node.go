package psi

import (
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

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

type Node interface {
	ID() int64
	UUID() string
	Node() *NodeBase

	Parent() Node
	SetParent(parent Node)
	Children() []Node

	Edges() []Edge

	IsContainer() bool
	IsLeaf() bool

	Comments() []string

	attachToGraph(g *Graph)
	detachFromGraph(g *Graph)

	addChildNode(node Node)
	removeChildNode(node Node)
	replaceChildNode(old Node, node Node)

	setEdge(edge Edge)
	unsetEdge(key EdgeKey)
	getEdge(key EdgeKey) Edge
}
type NodeBase struct {
	g        *Graph
	id       int64
	uuid     string
	self     Node
	parent   Node
	children []Node
	edges    map[EdgeKey]Edge
}

func (n *NodeBase) Init(self Node, uid string) {
	n.self = self
	n.uuid = uid
	n.edges = map[EdgeKey]Edge{}

	if n.uuid == "" {
		n.uuid = uuid.New().String()
	}
}

func (n *NodeBase) ID() int64       { return n.id }
func (n *NodeBase) UUID() string    { return n.uuid }
func (n *NodeBase) Node() *NodeBase { return n }
func (n *NodeBase) Parent() Node    { return n.parent }

func (n *NodeBase) Children() []Node { return n.children }
func (n *NodeBase) Edges() []Edge    { return nil }

func (n *NodeBase) SetParent(parent Node) {
	// TODO: Write Godoc for this method
	if n.parent == parent {
		return
	}

	if n.parent != nil {
		n.parent.removeChildNode(n.self)
		n.parent = nil
	}

	n.parent = parent

	if n.parent != nil {
		n.parent.addChildNode(n.self)
	} else {
		n.detachFromGraph(nil)
	}
}

func (n *NodeBase) setEdge(edge Edge) {
	k := edge.Key()

	if edge.From() != n.self {
		panic("invalid edge")
	}

	if e, ok := n.edges[k]; ok {
		if e.To() == edge.To() {
			return
		}

		n.edges[k] = edge.ReplaceTo(n.self)
	} else {
		n.edges[k] = edge
	}
}

func (n *NodeBase) unsetEdge(key EdgeKey)    { delete(n.edges, key) }
func (n *NodeBase) getEdge(key EdgeKey) Edge { return n.edges[key] }

func (n *NodeBase) addChildNode(child Node) {
	idx := slices.Index(n.children, child)

	if idx != -1 {
		return
	}

	n.children = append(n.children, child)

	child.attachToGraph(n.g)
}

func (n *NodeBase) removeChildNode(child Node) {
	idx := slices.Index(n.children, child)

	if idx == -1 {
		return
	}

	n.children = slices.Delete(n.children, idx, idx+1)
}

func (n *NodeBase) replaceChildNode(old, new Node) {
	idx := slices.Index(n.children, old)

	if idx != -1 {
		old.SetParent(nil)
		n.children[idx] = new
		new.SetParent(n.self)
	}

	for i, e := range n.edges {
		if e.To() == old {
			e = e.ReplaceTo(new)
		} else {
			continue
		}

		n.edges[i] = e
	}
}

func (n *NodeBase) attachToGraph(g *Graph) {
	if n.g == g {
		return
	}

	if g == nil {
		n.detachFromGraph(nil)
		return
	}

	if n.g != nil {
		panic("node already attached to a graph")
	}

	n.g = g
	n.id = g.g.NewNode().ID()

	for _, e := range n.children {
		e.attachToGraph(g)
	}
}

func (n *NodeBase) detachFromGraph(g *Graph) {
	if n.g == nil {
		return
	}

	if n.g != g {
		return
	}

	for _, e := range n.children {
		e.detachFromGraph(n.g)
	}

	n.g = nil
}
