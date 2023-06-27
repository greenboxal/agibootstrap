package psi

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/platform/obsfx/collectionsfx"
)

var EdgeKindChild = EdgeKind("child")

type NodeID = string

type InvalidationListener interface {
	OnInvalidated(n Node)
}

type invalidationListenerFunc struct{ f func(n Node) }

func InvalidationListenerFunc(f func(n Node)) InvalidationListener {
	return &invalidationListenerFunc{f: f}
}

func (f *invalidationListenerFunc) OnInvalidated(n Node) { f.f(n) }

type NodeLike interface {
	PsiNode() Node
	PsiNodeBase() *NodeBase
}

type NodeIterator interface {
	Value() Node
	Node() Node
	Next() bool
}

type NamedNode interface {
	Node

	PsiNodeName() string
}

// Node represents a PSI element in the graph.
type Node interface {
	NodeLike

	ID() int64
	UUID() NodeID

	Parent() Node
	PreviousSibling() Node
	NextSibling() Node
	CanonicalPath() Path

	// SetParent sets the parent node of the current node.
	// If the parent node is already set to the given parent, no action is taken.
	// If the current node has a parent, it is first removed from its parent node.
	// Then, the parent node is set to the given parent.
	// If the parent node is not nil, the current node is added as a child to the parent node.
	// If the parent node is nil, the current node is detached from the graph.
	SetParent(parent Node)

	Children() []Node
	ChildrenList() collectionsfx.ObservableList[Node]
	ChildrenIterator() NodeIterator
	Edges() EdgeIterator
	Comments() []string

	IsContainer() bool
	IsLeaf() bool

	ResolveChild(component PathElement) Node

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
	SetEdge(key EdgeReference, to Node)
	UnsetEdge(key EdgeReference)
	GetEdge(key EdgeReference) Edge

	SetAttribute(key string, value any)
	GetAttribute(key string) (any, bool)
	RemoveAttribute(key string) (any, bool)

	IsValid() bool
	Invalidate()
	Update()

	AddInvalidationListener(listener InvalidationListener)
	RemoveInvalidationListener(listener InvalidationListener)

	attachToGraph(g Graph)
	detachFromGraph(g Graph)

	AddChildNode(node Node)
	RemoveChildNode(node Node)
	ReplaceChildNode(old Node, node Node)
	InsertChildrenAt(idx int, child Node)
	InsertChildBefore(anchor Node, node Node)
	InsertChildAfter(anchor Node, node Node)

	String() string
}

type NodeLikeBase struct {
	NodeBase NodeBase
}

func (n *NodeLikeBase) PsiNode() Node          { return n.NodeBase.PsiNode() }
func (n *NodeLikeBase) PsiNodeBase() *NodeBase { return n.NodeBase.PsiNodeBase() }

type NodeBase struct {
	g Graph

	id      int64
	uuid    string
	version int64

	parent Node
	self   Node
	path   Path

	children   collectionsfx.MutableSlice[Node]
	edges      collectionsfx.MutableMap[EdgeKey, Edge]
	attributes collectionsfx.MutableMap[string, any]

	valid                 bool
	inUpdate              bool
	invalidationListeners []InvalidationListener
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

func (n *NodeBase) PsiNode() Node          { return n.self }
func (n *NodeBase) PsiNodeBase() *NodeBase { return n }

func (n *NodeBase) ID() int64          { return n.id }
func (n *NodeBase) UUID() string       { return n.uuid }
func (n *NodeBase) IsContainer() bool  { return true }
func (n *NodeBase) IsLeaf() bool       { return false }
func (n *NodeBase) IsValid() bool      { return n.valid }
func (n *NodeBase) Comments() []string { return nil }

func (n *NodeBase) CanonicalPath() (res Path)                        { return n.path }
func (n *NodeBase) Parent() Node                                     { return n.parent }
func (n *NodeBase) Children() []Node                                 { return n.children.Slice() }
func (n *NodeBase) ChildrenList() collectionsfx.ObservableList[Node] { return &n.children }
func (n *NodeBase) ChildrenIterator() NodeIterator                   { return &nodeChildrenIterator{parent: n} }

func (n *NodeBase) String() string {
	return fmt.Sprintf("Node(%T, %d, %s)", n.self, n.id, n.uuid)
}

func (n *NodeBase) AddInvalidationListener(listener InvalidationListener) {
	if slices.Index(n.invalidationListeners, listener) != -1 {
		return
	}

	n.invalidationListeners = append(n.invalidationListeners, listener)
}

func (n *NodeBase) RemoveInvalidationListener(listener InvalidationListener) {
	for i, l := range n.invalidationListeners {
		if l == listener {
			n.invalidationListeners = append(n.invalidationListeners[:i], n.invalidationListeners[i+1:]...)
			return
		}
	}
}

func (n *NodeBase) Invalidate() {
	if !n.valid {
		n.valid = false

		if n.parent != nil {
			n.parent.Invalidate()
		}
	}
}

func (n *NodeBase) Update() {
	if !n.inUpdate {
		n.doUpdate(false)
	}
}

func (n *NodeBase) doUpdate(skipValidation bool) {
	if n.valid && !skipValidation {
		return
	}

	n.inUpdate = true

	defer func() {
		n.inUpdate = false
	}()

	n.updatePath()

	for it := n.children.Iterator(); it.Next(); {
		it.Item().PsiNodeBase().doUpdate(skipValidation)
	}

	n.Update()

	n.valid = true

	n.fireInvalidationListeners()
}

func (n *NodeBase) fireInvalidationListeners() {
	for _, listener := range n.invalidationListeners {
		listener.OnInvalidated(n)
	}
}
func (n *NodeBase) updatePath() {
	var self PathElement

	if n.parent == nil {
		self.Kind = EdgeKindChild
		self.Name = n.UUID()
		n.path = PathFromComponents(self)
		return
	}

	parentPath := n.parent.CanonicalPath()

	if named, ok := n.self.(NamedNode); ok {
		self = PathElement{
			Kind: EdgeKindChild,
			Name: named.PsiNodeName(),
		}
	} else {
		index := n.parent.PsiNodeBase().IndexOfChild(n.self)

		self = PathElement{
			Kind:  EdgeKindChild,
			Index: int64(index),
		}
	}

	n.path = parentPath.Child(self)
}

func (n *NodeBase) ResolveChild(component PathElement) Node {
	if component.Kind == "" || component.Kind == EdgeKindChild {
		if component.Name == "" {
			if component.Index < int64(n.children.Len()) {
				return n.children.Get(int(component.Index))
			}
		} else {
			for it := n.children.Iterator(); it.Next(); {
				child := it.Item()
				cn := child.PsiNodeBase()

				if named, ok := child.(NamedNode); ok {
					if named.PsiNodeName() == component.Name {
						return child
					}
				}

				if component.Name == cn.UUID() {
					return child
				}
			}
		}
	} else {
		for it := n.edges.Iterator(); it.Next(); {
			kv := it.Item()
			k := kv.Key

			if k.GetKind() != component.Kind {
				continue
			}

			if k.GetName() != component.Name {
				continue
			}

			if k.GetIndex() != component.Index {
				continue
			}

			return kv.Value.To()
		}
	}

	return nil
}
func (n *NodeBase) PreviousSibling() Node {
	if n.parent == nil {
		return nil
	}

	p := n.Parent().PsiNodeBase()
	idx := p.children.IndexOf(n.self)

	if idx <= 0 {
		return nil
	}

	return p.children.Get(idx - 1)
}

func (n *NodeBase) NextSibling() Node {
	if n.parent == nil {
		return nil
	}

	p := n.Parent().PsiNodeBase()
	idx := p.children.IndexOf(n.self)

	if idx == -1 || idx >= p.children.Len()-1 {
		return nil
	}

	return p.children.Get(idx + 1)
}

func (n *NodeBase) Edges() EdgeIterator {
	return &edgeIterator{
		n:    n,
		keys: n.edges.Keys(),
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
		n.parent.RemoveChildNode(n.self)
		n.parent = nil
	}

	n.parent = parent

	if n.parent != nil {
		n.parent.AddChildNode(n.self)
	} else {
		n.detachFromGraph(nil)
	}

	n.doUpdate(true)
}

// AddChildNode adds a child node to the current node.
// If the child node is already a child of the current node, no action is taken.
// The child node is appended to the list of children nodes of the current node.
// Then, the child node is attached to the same graph as the parent node.
//
// Parameters:
// - child: The child node to be added.
func (n *NodeBase) AddChildNode(child Node) {
	n.InsertChildrenAt(n.children.Len(), child)
}

// RemoveChildNode removes the child node from the current node.
// If the child node is not a child of the current node, no action is taken.
//
// Parameters:
// - child: The child node to be removed.
func (n *NodeBase) RemoveChildNode(child Node) {
	idx := n.children.IndexOf(child)

	if idx == -1 {
		return
	}

	n.children.RemoveAt(idx)

	n.Invalidate()
	n.fireInvalidationListeners()
}

func (n *NodeBase) InsertChildrenAt(idx int, child Node) {
	if child == n || child == n.self {
		panic("invalid child")
	}

	cn := child.PsiNodeBase()

	existingIdx := n.children.IndexOf(child)

	if existingIdx != -1 && idx == existingIdx {
		return
	}

	cn.parent = n
	n.children.InsertAt(idx, child)

	if existingIdx != -1 {
		if existingIdx >= idx {
			existingIdx++
		}

		n.children.RemoveAt(existingIdx)
	}

	child.attachToGraph(n.g)

	n.Invalidate()
	n.fireInvalidationListeners()
}

func (n *NodeBase) InsertChildBefore(anchor, node Node) {
	idx := n.children.IndexOf(anchor)

	if idx == -1 {
		return
	}

	n.InsertChildrenAt(idx, node)
}

func (n *NodeBase) InsertChildAfter(anchor, node Node) {
	idx := n.children.IndexOf(anchor)

	if idx == -1 {
		return
	}

	n.InsertChildrenAt(idx+1, node)
}

// ReplaceChildNode replaces an old child node with a new child node in the current node.
// If the old child node is not a child of the current node, no action is taken.
// The old child node is first removed from its parent node and detached from the graph.
// Then, the new child node is set as the replacement at the same index in the list of children nodes of the current node.
// The new child node is attached to the same graph as the parent node.
// Finally, any edges in the current node that reference the old child node as the destination node are updated to reference the new child node.
//
// Parameters:
// - old: The old child node to be replaced.
// - new: The new child node to replace the old child node.
func (n *NodeBase) ReplaceChildNode(old, new Node) {
	changed := false
	idx := n.children.IndexOf(old)

	if idx != -1 {
		n.children.Set(idx, new)
		old.SetParent(nil)
		new.SetParent(n.self)

		changed = true
	}

	for it := n.edges.Iterator(); it.Next(); {
		kv := it.Item()
		e := kv.Value

		if e.To() == old {
			e = e.ReplaceTo(new)
		} else {
			continue
		}

		n.edges.Set(kv.Key, e)

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
func (n *NodeBase) attachToGraph(g Graph) {
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
	n.id = g.AllocateNodeID()

	if n.g != nil {
		n.g.Add(n.self)
	}

	for it := n.children.Iterator(); it.Next(); {
		it.Item().attachToGraph(g)
	}

	for it := n.edges.Iterator(); it.Next(); {
		it.Item().Value.attachToGraph(g)
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
func (n *NodeBase) detachFromGraph(g Graph) {
	if n.g == nil {
		return
	}

	if n.g != g {
		return
	}

	for it := n.children.Iterator(); it.Next(); {
		it.Item().detachFromGraph(n.g)
	}

	oldGraph := n.g

	n.g = nil

	oldGraph.Remove(n.self)

	n.Invalidate()
}

func (n *NodeBase) SetAttribute(key string, value any) {
	n.attributes.Set(key, value)

	n.Invalidate()
}

func (n *NodeBase) GetAttribute(key string) (value any, ok bool) {
	return n.attributes.Get(key)
}

func (n *NodeBase) RemoveAttribute(key string) (value any, ok bool) {
	value, ok = n.attributes.Get(key)

	if !ok {
		return value, false
	}

	n.attributes.Remove(key)

	if ok {
		n.Invalidate()
	}

	return
}

func (n *NodeBase) SetEdge(key EdgeReference, to Node) {
	e := &EdgeBase{
		from: n.self,
		to:   to,
		key:  key,
	}

	n.edges.Set(e.key.GetKey(), e)

	n.doUpdate(true)
}

func (n *NodeBase) UnsetEdge(key EdgeReference) {
	k := key.GetKey()

	_, ok := n.edges.Get(k)

	if !ok {
		return
	}

	if ok {
		n.Invalidate()
	}
}
func (n *NodeBase) GetEdge(key EdgeReference) Edge {
	v, _ := n.edges.Get(key.GetKey())

	return v
}

func (n *NodeBase) IndexOfChild(node Node) int {
	return n.children.IndexOf(node)
}

type nodeSliceIterator struct {
	current Node
	items   []Node
}

func (n *nodeSliceIterator) Value() Node { return n.Node() }

func (n *nodeSliceIterator) Node() Node {
	return n.current
}

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

func (n *nodeChildrenIterator) Value() Node { return n.Node() }

func (n *nodeChildrenIterator) Node() Node {
	return n.current
}

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

func (n *nestedNodeIterator) Value() Node { return n.Node() }

func (n *nestedNodeIterator) Node() Node {
	if n.current == nil {
		return nil
	}

	return n.current.Node()
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

func AppendNodeIterator(iterators ...NodeIterator) NodeIterator {
	return &nestedNodeIterator{iterators: iterators}
}
