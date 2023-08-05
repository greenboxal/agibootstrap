package psi

import (
	"context"
	"fmt"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
)

type NodeBase struct {
	g    Graph
	snap NodeSnapshot

	typ NodeType

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

	n.initialize()
}

func (n *NodeBase) initialize() {
	n.updatePath()

	if n.snap != nil {
		n.snap.OnBeforeInitialize(n.self)
	}

	obsfx.ObserveChange(&n.parent, n.onParentChange)
	n.children.AddListListener(collectionsfx.OnListChangedFunc[Node](n.onChildrenChange))
	n.edges.AddMapListener(collectionsfx.OnMapChangedFunc[EdgeKey, Edge](n.onEdgesChange))
	n.attributes.AddMapListener(collectionsfx.OnMapChangedFunc[string, any](n.onAttributeChange))

	if onInit, ok := n.typ.(baseNodeOnInitialize); ok {
		onInit.OnInitNode(n.self)
	}

	if n.snap != nil {
		n.snap.OnAfterInitialize(n.self)
	}
}

func (n *NodeBase) ID() int64 {
	if n.snap != nil {
		return n.snap.ID()
	}

	return -1
}

func (n *NodeBase) PsiNodeVersion() int64 {
	if n.snap == nil {
		return 0
	}

	return n.snap.CommitVersion()
}

func (n *NodeBase) Graph() Graph { return n.g }

func (n *NodeBase) SetSnapshot(snapshot NodeSnapshot) { n.snap = snapshot }
func (n *NodeBase) GetSnapshot() NodeSnapshot         { return n.snap }

func (n *NodeBase) PsiNode() Node          { return n.self }
func (n *NodeBase) PsiNodeBase() *NodeBase { return n }
func (n *NodeBase) PsiNodeType() NodeType  { return n.typ }

func (n *NodeBase) IsContainer() bool  { return true }
func (n *NodeBase) IsLeaf() bool       { return false }
func (n *NodeBase) IsValid() bool      { return n.valid }
func (n *NodeBase) Comments() []string { return nil }

func (n *NodeBase) CanonicalPath() Path {
	if n.snap != nil {
		return n.snap.Path()
	}

	return n.path
}

func (n *NodeBase) String() string {
	return fmt.Sprintf("Value(%T, %d, %s)", n.self, n.ID(), n.path)
}

func (n *NodeBase) ResolveChild(ctx context.Context, key PathElement) Node {
	if key.Kind == "" || key.Kind == EdgeKindChild {
		if key.Name == "" {
			if key.Index < int64(n.children.Len()) {
				return n.children.Get(int(key.Index))
			}
		} else {
			for it := n.children.Iterator(); it.Next(); {
				child := it.Item()

				if named, ok := child.(NamedNode); ok {
					if named.PsiNodeName() == key.Name {
						return child
					}
				}
			}
		}
	} else {
		for it := n.edges.Iterator(); it.Next(); {
			kv := it.Item()
			k := kv.Key

			if k.GetKind() != key.Kind {
				continue
			}

			if k.GetName() != key.Name {
				continue
			}

			if k.GetIndex() != key.Index {
				continue
			}

			return kv.Value.To()
		}
	}

	if n.snap != nil {
		if resolved, err := n.snap.Resolve(ctx, PathFromElements("", true, key)); err == nil {
			return resolved
		}
	}

	typ := key.Kind.Type()

	if typ == nil {
		return nil
	}

	resolved, err := typ.Resolve(ctx, n.g, n.self, key.AsEdgeKey())

	if err != nil {
		return nil
	}

	return resolved
}

func (n *NodeBase) Update(ctx context.Context) error {
	if n.valid {
		return nil
	}

	n.inUpdate = true

	defer func() {
		n.inUpdate = false
	}()

	if n, ok := n.PsiNode().(UpdatableNode); ok {
		if err := n.OnUpdate(ctx); err != nil {
			return nil
		}
	}

	n.valid = true

	if n.snap != nil {
		if err := n.snap.OnUpdated(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (n *NodeBase) OnUpdate(ctx context.Context) error {
	for it := n.children.Iterator(); it.Next(); {
		if err := it.Item().PsiNodeBase().Update(ctx); err != nil {
			return err
		}
	}

	return nil
}
func (n *NodeBase) SelfIdentity() Path {
	var self PathElement

	if n.Parent() == nil {
		if unique, ok := n.self.(UniqueNode); ok {
			return PathFromElements(unique.UUID(), false)
		}
	}

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

	return PathFromElements("", true, self)
}

func (n *NodeBase) updatePath() {
	self := n.SelfIdentity()

	if p := n.parent.Value(); p != nil {
		self = p.CanonicalPath().Join(self)
	}

	n.path = self

	for it := n.children.Iterator(); it.Next(); {
		it.Item().PsiNodeBase().updatePath()
	}
}

func (n *NodeBase) Invalidate() {
	if n.valid {
		return
	}

	n.valid = false

	if n.snap != nil {
		n.snap.OnInvalidated()
	}
}

func (n *NodeBase) InvalidateTree() {
	n.Invalidate()

	for it := n.ChildrenIterator(); it.Next(); {
		it.Value().PsiNodeBase().InvalidateTree()
	}
}

func (n *NodeBase) AttachToGraph(g Graph) {
	if n.g == g {
		return
	}

	if n.g != nil {
		panic("node already attached to a graph")
	}

	n.g = g

	if n.g != nil {
		n.g.Add(n.self)
	}

	for it := n.children.Iterator(); it.Next(); {
		it.Item().PsiNodeBase().AttachToGraph(g)
	}

	for it := n.edges.Iterator(); it.Next(); {
		it.Item().Value.attachToGraph(g)
	}

	n.Invalidate()
}

func (n *NodeBase) DetachFromGraph(g Graph) {
	if n.g == nil || n.g != g {
		return
	}

	for it := n.children.Iterator(); it.Next(); {
		it.Item().PsiNodeBase().DetachFromGraph(n.g)
	}

	oldGraph := n.g

	n.g = nil

	oldGraph.Remove(n.self)

	n.Invalidate()
}

func (n *NodeBase) onParentChange(old, new Node) {
	if old != nil {
		old.PsiNodeBase().RemoveChildNode(n.self)
	}

	if new != nil {
		new.PsiNodeBase().AddChildNode(n.self)

		n.updatePath()

		n.AttachToGraph(new.PsiNodeBase().g)
	} else {
		n.updatePath()
	}

	n.InvalidateTree()

	if n.snap != nil {
		n.snap.OnParentChange(new)
	}
}

func (n *NodeBase) onChildrenChange(ev collectionsfx.ListChangeEvent[Node]) {
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

					child.SetParent(n.self)
					child.PsiNodeBase().indexInParent = ev.From() + i
				}
			}
		}
	}

	n.Invalidate()
}

func (n *NodeBase) onEdgesChange(ev collectionsfx.MapChangeEvent[EdgeKey, Edge]) {
	if ev.WasAdded {
		ev.ValueAdded.attachToGraph(n.g)
	}

	if n.snap != nil {
		if ev.WasAdded {
			n.snap.OnEdgeAdded(ev.ValueAdded)
		} else if ev.WasRemoved {
			n.snap.OnEdgeRemoved(ev.ValueRemoved)
		}
	}

	n.Invalidate()
}

func (n *NodeBase) onAttributeChange(ev collectionsfx.MapChangeEvent[string, any]) {
	if n.snap != nil {
		if ev.WasAdded {
			n.snap.OnAttributeChanged(ev.Key, ev.ValueAdded)
		} else if ev.WasRemoved {
			n.snap.OnAttributeRemoved(ev.Key, ev.ValueRemoved)
		}
	}

	n.Invalidate()
}
