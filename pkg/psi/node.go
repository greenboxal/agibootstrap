package psi

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// Node represents a PSI element in the graph.
type Node interface {
	ID() int64
	UUID() string
	Node() *NodeBase

	Parent() Node
	CanonicalPath() Path

	// SetParent sets the parent node of the current node.
	// If the parent node is already set to the given parent, no action is taken.
	// If the current node has a parent, it is first removed from its parent node.
	// Then, the parent node is set to the given parent.
	// If the parent node is not nil, the current node is added as a child to the parent node.
	// If the parent node is nil, the current node is detached from the graph.
	SetParent(parent Node)

	Children() []Node
	Edges() EdgeIterator
	Comments() []string

	IsContainer() bool
	IsLeaf() bool

	// SetEdge sets the given edge on the current node.
	// It checks if the edge is valid by verifying that it originates from the current node.
	// If an edge with the same key already exists, it replaces the edge's destination node with the current node.
	// If the edge is not found, it adds the edge to the node's edge map.
	//
	// Parameters:
	// - edge: The edge to be set on the node.
	//
	// Panics:
	// - "invalid edge": If the given edge does not originate from the current node.
	SetEdge(key EdgeKey, to Node)
	UnsetEdge(key EdgeKey)
	GetEdge(key EdgeKey) Edge

	SetAttribute(key string, value any)
	GetAttribute(key string) (any, bool)
	RemoveAttribute(key string) (any, bool)

	IsValid() bool
	Invalidate()
	Update()

	attachToGraph(g *Graph)
	detachFromGraph(g *Graph)

	addChildNode(node Node)
	removeChildNode(node Node)
	replaceChildNode(old Node, node Node)
	insertChildNodeBefore(anchor Node, node Node)
	insertChildNodeAfter(anchor Node, node Node)

	String() string
}
type NodeBase struct {
	g *Graph

	id      int64
	uuid    string
	version int64

	parent Node
	self   Node
	path   Path

	children   []Node
	edges      map[EdgeKey]Edge
	attributes map[string]any

	valid bool
}

// Init initializes the NodeBase struct with the given self node and uid string.
// It sets the self node, uuid, and initializes the edges map.
// If the uuid is an empty string, it generates a new UUID using the github.com/google/uuid package.
//
// Parameters:
// - self: The self node to be set.
// - uid: The UUID string to be set.
func (n *NodeBase) Init(self Node, uid string) {
	n.self = self
	n.uuid = uid

	if n.uuid == "" {
		n.uuid = uuid.New().String()
	}
}

func (n *NodeBase) ID() int64          { return n.id }
func (n *NodeBase) UUID() string       { return n.uuid }
func (n *NodeBase) Node() *NodeBase    { return n }
func (n *NodeBase) Parent() Node       { return n.parent }
func (n *NodeBase) Children() []Node   { return n.children }
func (n *NodeBase) Comments() []string { return nil }
func (n *NodeBase) CanonicalPath() (res Path) {
	if n.parent != nil {
		res = append(res, n.parent.CanonicalPath()...)
	}

	res = append(res, n.path...)

	return
}
func (n *NodeBase) IsContainer() bool { return true }
func (n *NodeBase) IsLeaf() bool      { return false }

func (n *NodeBase) String() string {
	return fmt.Sprintf("Node(%T, %d, %s)", n.self, n.id, n.uuid)
}

func (n *NodeBase) IsValid() bool { return n.valid }

func (n *NodeBase) Invalidate() {
	if !n.valid {
		n.valid = false

		if n.parent != nil {
			n.parent.Invalidate()
		}
	}
}

func (n *NodeBase) Update() {
	if n.valid {
		return
	}

	for _, child := range n.children {
		child.Update()
	}

	n.valid = true
}

func (n *NodeBase) Edges() EdgeIterator {
	return &edgeIterator{
		n:    n,
		keys: maps.Keys(n.edges),
	}
}

func (n *NodeBase) SetParent(parent Node) {
	if parent == n || parent == n.self {
		panic("invalid parent")
	}

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

	n.Invalidate()
}

// addChildNode adds a child node to the current node.
// If the child node is already a child of the current node, no action is taken.
// The child node is appended to the list of children nodes of the current node.
// Then, the child node is attached to the same graph as the parent node.
//
// Parameters:
// - child: The child node to be added.
func (n *NodeBase) addChildNode(child Node) {
	n.insertChildrenAt(len(n.children), child)
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

	n.Invalidate()
}

func (n *NodeBase) insertChildrenAt(idx int, child Node) {
	if child == n || child == n.self {
		panic("invalid child")
	}

	cn := child.Node()

	existingIdx := slices.Index(n.children, child)

	if existingIdx != -1 && idx == existingIdx {
		return
	}

	cn.parent = n
	n.children = slices.Insert(n.children, idx, child)

	if existingIdx != -1 {
		if existingIdx >= idx {
			existingIdx++
		}

		n.children = slices.Delete(n.children, existingIdx, existingIdx+1)
	}

	cn.path = nil
	cn.path = append(cn.path, PathElement{
		Index: idx,
	})

	child.attachToGraph(n.g)

	n.Invalidate()
}

func (n *NodeBase) insertChildNodeBefore(anchor, node Node) {
	idx := slices.Index(n.children, anchor)

	if idx == -1 {
		return
	}

	n.insertChildrenAt(idx, node)
}

func (n *NodeBase) insertChildNodeAfter(anchor, node Node) {
	idx := slices.Index(n.children, anchor)

	if idx == -1 {
		return
	}

	n.insertChildrenAt(idx+1, node)
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
	changed := false
	idx := slices.Index(n.children, old)

	if idx != -1 {
		n.children[idx] = new
		old.SetParent(nil)
		new.SetParent(n.self)

		changed = true
	}

	for i, e := range n.edges {
		if e.To() == old {
			e = e.ReplaceTo(new)
		} else {
			continue
		}

		n.edges[i] = e

		changed = true
	}

	if changed {
		n.Invalidate()
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

	n.Invalidate()
}

// DetachFromGraph detaches the node from the given graph.
// If the node is not attached to any graph or if it is attached to a different graph,
// no action is taken.
// If the node is attached to the given graph, it is detached from the graph and all its
// children are recursively detached as well.
//
// Parameters:
// - g: The graph from which the node is to be detached.
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

	n.Invalidate()
}

func (n *NodeBase) SetAttribute(key string, value any) {
	if n.attributes == nil {
		n.attributes = make(map[string]any)
	}

	n.attributes[key] = value

	n.Invalidate()
}

func (n *NodeBase) GetAttribute(key string) (value any, ok bool) {
	if n.attributes == nil {
		return nil, false
	}

	value, ok = n.attributes[key]

	return
}

func (n *NodeBase) RemoveAttribute(key string) (value any, ok bool) {
	if n.attributes == nil {
		return nil, false
	}

	value, ok = n.attributes[key]

	delete(n.attributes, key)

	if ok {
		n.Invalidate()
	}

	return
}

func (n *NodeBase) SetEdge(key EdgeKey, to Node) {
	if n.edges == nil {
		n.edges = make(map[EdgeKey]Edge)
	}

	e := &EdgeBase{
		from: n.self,
		to:   to,
		key:  key,
	}

	n.edges[e.key] = e

	n.Invalidate()
}

func (n *NodeBase) UnsetEdge(key EdgeKey) {
	if n.edges == nil {
		return
	}

	_, ok := n.edges[key]

	delete(n.edges, key)

	if ok {
		n.Invalidate()
	}
}
func (n *NodeBase) GetEdge(key EdgeKey) Edge {
	if n.edges == nil {
		return nil
	}

	return n.edges[key]
}
