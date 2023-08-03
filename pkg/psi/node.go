package psi

import (
	"context"

	collectionsfx "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
)

type NodeID = string

type NodeLike interface {
	PsiNode() Node
	PsiNodeType() NodeType
	PsiNodeBase() *NodeBase
	PsiNodeVersion() int64
}

// Node represents a PSI element in the graph.
type Node interface {
	NodeLike

	ID() int64
	SelfIdentity() Path
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

	ResolveChild(ctx context.Context, component PathElement) Node

	/* Edges */

	// Edges returns the edges of the current node.
	Edges() EdgeIterator
	// UpsertEdge upserts the given edge.
	UpsertEdge(edge Edge)
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

	String() string
}

type NamedNode interface {
	Node

	PsiNodeName() string
}

type UpdatableNode interface {
	Node

	OnUpdate(context.Context) error
}

type UniqueNode interface {
	Node

	UUID() string
}

type NodeInitOption func(*NodeBase)

func WithNodeType(typ NodeType) NodeInitOption {
	return func(n *NodeBase) {
		n.typ = typ
	}
}

type NodeLikeBase struct {
	NodeBase NodeBase
}

func (n *NodeLikeBase) PsiNode() Node          { return n.NodeBase.PsiNode() }
func (n *NodeLikeBase) PsiNodeType() NodeType  { return n.NodeBase.PsiNodeType() }
func (n *NodeLikeBase) PsiNodeBase() *NodeBase { return n.NodeBase.PsiNodeBase() }
func (n *NodeLikeBase) PsiNodeVersion() int64  { return n.NodeBase.PsiNodeVersion() }
