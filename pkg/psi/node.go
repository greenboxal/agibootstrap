package psi

import (
	"github.com/dave/dst"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

type Node interface {
	ID() int64
	UUID() string
	Node() *NodeBase
	Parent() Node
	SetParent(parent Node)
	Children() []Node

	Ast() dst.Node

	IsContainer() bool
	IsLeaf() bool

	Comments() []string

	attachToGraph(g *Graph)
	detachFromGraph(g *Graph)
	addChildNode(node Node)
	removeChildNode(node Node)
}
type NodeBase struct {
	g        *Graph
	id       int64
	uuid     string
	self     Node
	parent   Node
	children []Node
}

func (n *NodeBase) Initialize(self Node, uid string) {
	n.self = self
	n.uuid = uid

	if n.uuid == "" {
		n.uuid = uuid.New().String()
	}
}

func (n *NodeBase) ID() int64        { return n.id }
func (n *NodeBase) UUID() string     { return n.uuid }
func (n *NodeBase) Node() *NodeBase  { return n }
func (n *NodeBase) Parent() Node     { return n.parent }
func (n *NodeBase) NodeBase() Node   { return n.parent }
func (n *NodeBase) Children() []Node { return n.children }

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
	n.id = g.NewNode().ID()

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
