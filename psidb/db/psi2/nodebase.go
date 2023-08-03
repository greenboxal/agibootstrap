package psi2

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type NodeBaseFlags uint32

const (
	NodeBaseFlagsNone        NodeBaseFlags = 0
	NodeBaseFlagsInitialized NodeBaseFlags = 1 << iota
	NodeBaseFlagsRoot
	NodeBaseFlagsNew
	NodeBaseFlagsDeleted
	NodeBaseFlagsDirty
	NodeBaseFlagsParentDirty
)

type INodeBase interface {
	Node

	GetSnapshot() NodeSnapshot
	SetSnapshot(snapshot NodeSnapshot)

	Update(ctx context.Context) error
}

type NodeEmbedder struct {
	base *NodeBase
}

func (n *NodeEmbedder) PsiNodeBase() *NodeBase        { return n.base }
func (n *NodeEmbedder) PsiNodeSnapshot() NodeSnapshot { return n.base.snapshot }

type NodeBase struct {
	snapshot NodeSnapshot

	self  Node
	typ   psi.NodeType
	flags NodeBaseFlags

	parent   MutableNodeReference[Node]
	children collectionsfx.MutableSlice[Node]
}

func (n *NodeBase) Init(self Node, options ...psi.NodeInitOption) {
	n.self = self

	// TODO: Implement options and default type

	n.initialize()
}

func (n *NodeBase) initialize() {
	if n.flags&NodeBaseFlagsInitialized != 0 {
		return
	}

	if n.snapshot != nil {
		n.snapshot.OnBeforeInitialize()
	}

	n.parent.SetValue(NewRelativeNodeReference[Node](n.self, psi.RelativePathToParent))

	obsfx.ObserveChange(&n.parent, func(old, new NodeReference[Node]) {
		if oldParent := ResolveLoadedOrNil(old); oldParent != nil {
			oldParent.PsiNodeBase().RemoveChild(n.self)
		}

		if newParent := ResolveLoadedOrNil(new); newParent != nil {
			newParent.PsiNodeBase().AddChild(n.self)

			n.flags &= ^NodeBaseFlagsRoot
		} else {
			n.flags |= NodeBaseFlagsRoot
		}

		n.flags |= NodeBaseFlagsParentDirty
	})

	n.flags |= NodeBaseFlagsInitialized

	if n.snapshot != nil {
		n.snapshot.OnAfterInitialize()
	}
}

func (n *NodeBase) ID() int64 {
	if n.snapshot == nil {
		return -1
	}

	return n.snapshot.ID()
}

func (n *NodeBase) SelfIdentity() psi.PathElement {
	if unique, ok := n.self.(UniqueNode); ok {
		return psi.PathElement{Name: unique.UUID()}
	}

	return psi.PathElement{}
}

func (n *NodeBase) Path() psi.Path {
	if n.snapshot == nil {
		if p := n.parent.ResolveSync(); p != nil {
			return p.PsiNodeBase().Path().Child(n.SelfIdentity())
		}

		return psi.PathFromElements("", true, n.SelfIdentity())
	}

	return n.snapshot.Path()
}

func (n *NodeBase) PsiNodeBase() *NodeBase        { return n }
func (n *NodeBase) PsiNodeSnapshot() NodeSnapshot { return n.snapshot }

func (n *NodeBase) GetParent() Node {
	return n.parent.ResolveSync()
}

func (n *NodeBase) SetParent(parent Node) {
	n.parent.SetValue(NewStaticNodeReference(parent))
}

func (n *NodeBase) GetSnapshot() NodeSnapshot         { return n.snapshot }
func (n *NodeBase) SetSnapshot(snapshot NodeSnapshot) { n.snapshot = snapshot }

func (n *NodeBase) AddChild(self Node) {
	n.
}

func (n *NodeBase) RemoveChild(self Node) {

}
