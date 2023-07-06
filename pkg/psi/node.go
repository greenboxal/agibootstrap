package psi

import (
	"context"
	"fmt"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipld/go-ipld-prime"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	collectionsfx "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
)

type NodeID = string

type NodeLike interface {
	PsiNode() Node
	PsiNodeType() NodeType
	PsiNodeBase() *NodeBase
	PsiNodeVersion() int64
}

type NamedNode interface {
	Node

	PsiNodeName() string
}

type UpdateableNode interface {
	Node

	OnUpdate(context.Context) error
}

// Node represents a PSI element in the graph.
type Node interface {
	obsfx.Observable

	NodeLike

	ID() int64
	UUID() NodeID
	CanonicalPath() Path

	Parent() Node
	PreviousSibling() Node
	NextSibling() Node

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

	Comments() []string

	IsContainer() bool
	IsLeaf() bool

	ResolveChild(component PathElement) Node

	/* Edges */

	// Edges returns the edges of the current node.
	Edges() EdgeIterator
	// SetEdge sets the edge with the given key to the given node.
	SetEdge(key EdgeReference, to Node)
	// UnsetEdge removes the edge with the given key.
	UnsetEdge(key EdgeReference)
	// GetEdge returns the edge with the given key.
	GetEdge(key EdgeReference) Edge

	/* Attributes */

	// Attributes returns the attributes of the current node.
	Attributes() map[string]interface{}
	// SetAttribute sets the attribute with the given key to the given value.
	SetAttribute(key string, value any)
	// GetAttribute returns the attribute with the given key.
	GetAttribute(key string) (any, bool)
	// RemoveAttribute removes the attribute with the given key.
	RemoveAttribute(key string) (any, bool)

	IsValid() bool
	Invalidate()
	Update(context.Context) error

	AddChildNode(node Node)
	RemoveChildNode(node Node)
	ReplaceChildNode(old Node, node Node)
	InsertChildrenAt(idx int, child Node)
	InsertChildBefore(anchor Node, node Node)
	InsertChildAfter(anchor Node, node Node)

	getGraph() Graph
	attachToGraph(g Graph)
	detachFromGraph(g Graph)
	setLastSnapshot(snapshot *NodeSnapshot)
	getLastSnapshot() *NodeSnapshot

	String() string
}

type NodeLikeBase struct {
	NodeBase NodeBase
}

func (n *NodeLikeBase) PsiNode() Node          { return n.NodeBase.PsiNode() }
func (n *NodeLikeBase) PsiNodeType() NodeType  { return n.NodeBase.PsiNodeType() }
func (n *NodeLikeBase) PsiNodeBase() *NodeBase { return n.NodeBase.PsiNodeBase() }
func (n *NodeLikeBase) PsiNodeVersion() int64  { return n.NodeBase.PsiNodeVersion() }

type NodeBase struct {
	obsfx.HasListenersBase[obsfx.InvalidationListener]

	g            Graph
	lastSnapshot *NodeSnapshot

	id      int64
	typ     NodeType
	version int64

	self Node
	path Path

	parent        obsfx.SimpleProperty[Node]
	indexInParent int

	children   collectionsfx.MutableSlice[Node]
	edges      collectionsfx.MutableMap[EdgeKey, Edge]
	attributes collectionsfx.MutableMap[string, any]

	valid    bool
	inUpdate bool
}

func (n *NodeBase) getGraph() Graph                        { return n.g }
func (n *NodeBase) setLastSnapshot(snapshot *NodeSnapshot) { n.lastSnapshot = snapshot }
func (n *NodeBase) getLastSnapshot() *NodeSnapshot         { return n.lastSnapshot }

type NodeInitOption func(*NodeBase)

func WithNodeID(id int64) NodeInitOption {
	return func(n *NodeBase) {
		n.id = id
	}
}

func WithNodeType(typ NodeType) NodeInitOption {
	return func(n *NodeBase) {
		n.typ = typ
	}
}

func WithNodeVersion(version int64) NodeInitOption {
	return func(n *NodeBase) {
		n.version = version
	}
}

// Init initializes the NodeBase struct with the given self node and uid string.
// It sets the self node, uuid, and initializes the edges map.
// If the uuid is an empty string, it generates a new UUID using the github.com/google/uuid package.
//
// Parameters:
// - self: The self node to be set.
// - uid: The UUID string to be set.
func (n *NodeBase) Init(self Node, options ...NodeInitOption) {
	if self == nil {
		panic("self node cannot be nil")
	}

	if n.self != nil {
		panic(fmt.Sprintf("node %v already initialized", n.self))
	}

	n.self = self

	for _, option := range options {
		option(n)
	}

	if n.typ == nil {
		n.typ = ReflectNodeType(typesystem.TypeOf(self))
	}

	obsfx.ObserveChange(&n.parent, func(old, new Node) {
		if old != nil {
			old.PsiNodeBase().RemoveChildNode(n.self)
		}

		if new != nil {
			new.PsiNodeBase().AddChildNode(n.self)

			n.updatePath()

			n.attachToGraph(new.PsiNodeBase().g)
		} else {
			n.updatePath()
		}

		n.InvalidateTree()
	})

	collectionsfx.ObserveList(&n.children, func(ev collectionsfx.ListChangeEvent[Node]) {
		for ev.Next() {
			if ev.WasPermutated() {
				for i := ev.From(); i < ev.To(); i++ {
					u := ev.GetPermutation(i)

					a := n.children.Get(i)
					b := n.children.Get(u)

					a.PsiNodeBase().indexInParent = i
					b.PsiNodeBase().indexInParent = u
				}
			} else {
				if ev.WasRemoved() {
					for _, child := range ev.RemovedSlice() {
						if child.Parent() == n.self {
							child.SetParent(nil)
							child.PsiNodeBase().indexInParent = -1
						}
					}
				}

				if ev.WasAdded() {
					for i, child := range ev.AddedSlice() {
						if child == nil || child == n || child == n.self || n == child.PsiNodeBase() {
							panic("invalid child")
						}

						if child.Parent() != n.self {
							child.SetParent(n.self)
						}

						child.PsiNodeBase().indexInParent = ev.From() + i
					}
				}
			}
		}

		n.Invalidate()
	})

	collectionsfx.ObserveMap(&n.edges, func(ev collectionsfx.MapChangeEvent[EdgeKey, Edge]) {
		if ev.WasAdded {
			ev.ValueAdded.attachToGraph(n.g)
		}

		n.Invalidate()
	})

	collectionsfx.ObserveMap(&n.attributes, func(ev collectionsfx.MapChangeEvent[string, any]) {
		n.Invalidate()
	})
}

func (n *NodeBase) PsiNode() Node          { return n.self }
func (n *NodeBase) PsiNodeBase() *NodeBase { return n }
func (n *NodeBase) PsiNodeType() NodeType  { return n.typ }
func (n *NodeBase) PsiNodeVersion() int64  { return n.version }

func (n *NodeBase) PsiNodeLink() ipld.Link {
	if n.lastSnapshot == nil {
		return nil
	}

	return n.lastSnapshot.Link
}

func (n *NodeBase) ID() int64          { return n.id }
func (n *NodeBase) IsContainer() bool  { return true }
func (n *NodeBase) IsLeaf() bool       { return false }
func (n *NodeBase) IsValid() bool      { return n.valid }
func (n *NodeBase) Comments() []string { return nil }

func (n *NodeBase) CanonicalPath() Path                         { return n.path }
func (n *NodeBase) Parent() Node                                { return n.parent.Value() }
func (n *NodeBase) ParentProperty() obsfx.ObservableValue[Node] { return &n.parent }

func (n *NodeBase) Children() []Node                                 { return n.children.Slice() }
func (n *NodeBase) ChildrenList() collectionsfx.ObservableList[Node] { return &n.children }
func (n *NodeBase) ChildrenIterator() NodeIterator                   { return &nodeChildrenIterator{parent: n} }

func (n *NodeBase) String() string {
	return fmt.Sprintf("Value(%T, %d, %s)", n.self, n.id, n.path)
}

func (n *NodeBase) UUID() NodeID {
	return n.CanonicalPath().String()
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

				if named, ok := child.(NamedNode); ok {
					if named.PsiNodeName() == component.Name {
						return child
					}
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

func (n *NodeBase) IndexOfChild(node Node) int {
	return n.children.IndexOf(node)
}

func (n *NodeBase) PreviousSibling() Node {
	if n.Parent() == nil {
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
	if n.Parent() == nil {
		return nil
	}

	p := n.Parent().PsiNodeBase()
	idx := p.children.IndexOf(n.self)

	if idx == -1 || idx >= p.children.Len()-1 {
		return nil
	}

	return p.children.Get(idx + 1)
}

func (n *NodeBase) SetParent(parent Node) {
	if n == parent || n.self == parent || (parent != nil && n == parent.PsiNodeBase()) {
		panic("invalid parent (cycle)")
	}

	n.parent.SetValue(parent)
}

// AddChildNode adds a child node to the current node.
// If the child node is already a child of the current node, no action is taken.
// The child node is appended to the list of children nodes of the current node.
// Then, the child node is attached to the same graph as the parent node.
//
// Parameters:
// - child: The child node to be added.
func (n *NodeBase) AddChildNode(child Node) {
	existingIdx := n.children.IndexOf(child)

	if existingIdx != -1 {
		return
	}

	n.children.Add(child)
}

// RemoveChildNode removes the child node from the current node.
// If the child node is not a child of the current node, no action is taken.
//
// Parameters:
// - child: The child node to be removed.
func (n *NodeBase) RemoveChildNode(child Node) {
	n.children.Remove(child)
}

func (n *NodeBase) InsertChildrenAt(idx int, child Node) {
	existingIdx := n.children.IndexOf(child)

	if existingIdx != -1 && idx == existingIdx {
		return
	}

	if existingIdx != -1 {
		if existingIdx >= idx {
			existingIdx++
		}

		n.children.RemoveAt(existingIdx)
	}

	n.children.InsertAt(idx, child)
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
	idx := n.children.IndexOf(old)

	if idx != -1 {
		n.children.Set(idx, new)
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
	}
}

func (n *NodeBase) Attributes() map[string]interface{} {
	return n.attributes.Map()
}

func (n *NodeBase) SetAttribute(key string, value any) {
	n.attributes.Set(key, value)
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

	return
}

func (n *NodeBase) Edges() EdgeIterator {
	return &nodeEdgeIterator{
		n: n,
	}
}

func (n *NodeBase) SetEdge(key EdgeReference, to Node) {
	e, _ := n.edges.Get(key.GetKey())

	if e == nil {
		e = &EdgeBase{
			from: n.self,
			to:   to,
			key:  key,
		}
	} else {
		if e.To() == to {
			return
		}

		e = e.ReplaceTo(to)
	}

	n.edges.Set(e.Key().GetKey(), e)
}

func (n *NodeBase) UnsetEdge(key EdgeReference) {
	k := key.GetKey()

	_, ok := n.edges.Get(k)

	if !ok {
		return
	}

	n.edges.Remove(k)
}
func (n *NodeBase) GetEdge(key EdgeReference) Edge {
	v, _ := n.edges.Get(key.GetKey())

	return v
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

	if n.g != nil {
		n.id = g.NextNodeID()

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

func (n *NodeBase) OnUpdate(ctx context.Context) error {
	for it := n.children.Iterator(); it.Next(); {
		if err := it.Item().PsiNodeBase().Update(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (n *NodeBase) Update(ctx context.Context) error {
	if n.valid {
		return nil
	}

	n.inUpdate = true

	defer func() {
		n.inUpdate = false
	}()

	n.version++

	if n, ok := n.PsiNode().(UpdateableNode); ok {
		if err := n.OnUpdate(ctx); err != nil {
			return nil
		}
	}

	n.valid = true

	if n.g != nil {
		n.g.OnNodeUpdated(n.self)
	}

	return nil
}

func (n *NodeBase) updatePath() {
	var self PathElement

	if n.Parent() == nil {
		self.Kind = EdgeKindChild
		self.Name = ""
		n.path = PathFromElements(self)
		return
	}

	parentPath := n.Parent().CanonicalPath()

	if named, ok := n.self.(NamedNode); ok {
		self = PathElement{
			Kind: EdgeKindChild,
			Name: named.PsiNodeName(),
		}
	} else {
		self = PathElement{
			Kind:  EdgeKindChild,
			Index: int64(n.indexInParent),
		}
	}

	n.path = parentPath.Child(self)
}

func (n *NodeBase) InvalidateTree() {
	n.Invalidate()

	for it := n.ChildrenIterator(); it.Next(); {
		it.Value().PsiNodeBase().InvalidateTree()
	}
}

func (n *NodeBase) Invalidate() {
	if !n.valid {
		return
	}

	n.valid = false

	n.ForEachListener(func(l obsfx.InvalidationListener) bool {
		l.OnInvalidated(n.self)

		return true
	})

	if n.g != nil {
		n.g.OnNodeInvalidated(n.self)
	}

	if n.Parent() != nil {
		n.Parent().Invalidate()
	}
}
