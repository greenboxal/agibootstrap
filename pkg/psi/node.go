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

// SetParent sets the parent node of the current node.
// If the parent node is already set to the given parent, no action is taken.
// If the current node has a parent, it is first removed from its parent node.
// Then, the parent node is set to the given parent.
// If the parent node is not nil, the current node is added as a child to the parent node.
// If the parent node is nil, the current node is detached from the graph.
func (n *NodeBase) SetParent(parent Node) {
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

// setEdge sets the given edge on the current node.
// It checks if the edge is valid by verifying that it originates from the current node.
// If an edge with the same key already exists, it replaces the edge's destination node with the current node.
// If the edge is not found, it adds the edge to the node's edge map.
//
// Parameters:
// - edge: The edge to be set on the node.
//
// Panics:
// - "invalid edge": If the given edge does not originate from the current node.
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

// addChildNode adds a child node to the current node.
// If the child node is already a child of the current node, no action is taken.
// The child node is appended to the list of children nodes of the current node.
// Then, the child node is attached to the same graph as the parent node.
//
// Parameters:
// - child: The child node to be added.
func (n *NodeBase) addChildNode(child Node) {
	idx := slices.Index(n.children, child)

	if idx != -1 {
		return
	}

	n.children = append(n.children, child)

	child.attachToGraph(n.g)
}

// removeChildNode removes the child node from the current node.
// If the child node is not a child of the current node, no action is taken.
//
// Parameters:
// - child: The child node to be removed.
func (n *NodeBase) removeChildNode(child Node) {
	idx := slices.Index(n.children, child)

	if idx == -1 {
		return
	}

	n.children = slices.Delete(n.children, idx, idx+1)
}

// replaceChildNode replaces an old child node with a new child node in the current node.
// If the old child node is not a child of the current node, no action is taken.
// The old child node is first removed from its parent node and detached from the graph.
// Then, the new child node is set as the replacement at the same index in the list of children nodes of the current node.
// The new child node is attached to the same graph as the parent node.
// Finally, any edges in the current node that reference the old child node as the destination node are updated to reference the new child node.
//
// Parameters:
// - old: The old child node to be replaced.
// - new: The new child node to replace the old child node.
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

// attachToGraph attaches the node to the given graph.
// If the node is already attached to the given graph, no action is taken.
// If the graph is nil, the node is detached from its current graph.
// If the node is already attached to a different graph, it raises a panic.
// The node is assigned a new ID from the graph's NewNode method.
// After attaching the node, each child of the node is also attached to the graph recursively.
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
	// TODO: Write Godoc for this method.
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

// DetachFromGraph detaches the node from the given graph.
// If the node is not attached to any graph or if it is attached to a different graph,
// no action is taken.
// If the node is attached to the given graph, it is detached from the graph and all its
// children are recursively detached as well.
//
// Parameters:
// - g: The graph from which the node is to be detached.
func (n *NodeBase) DetachFromGraph(g *Graph) {
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
