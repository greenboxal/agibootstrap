package psi

import (
	"github.com/dave/dst"
	"golang.org/x/exp/slices"
)

type Node interface {
	ID() int64
	UUID() string
	Node() dst.Node
	Parent() *Container
	Children() []Node

	IsContainer() bool
	IsLeaf() bool

	Comments() []string

	attachToGraph(g *Graph)
	detachFromGraph(g *Graph)
	setParent(parent *Container)
}
type nodeBase struct {
	g        *Graph
	id       int64
	uuid     string
	self     Node
	parent   *Container
	children []Node
}

func (n *nodeBase) ID() int64          { return n.id }
func (n *nodeBase) UUID() string       { return n.uuid }
func (n *nodeBase) Parent() *Container { return n.parent }
func (n *nodeBase) Children() []Node   { return n.children }

func (n *nodeBase) addChildNode(child Node) {
	idx := slices.Index(n.children, child)

	if idx != -1 {
		return
	}

	n.children = append(n.children, child)

	child.attachToGraph(n.g)
}

func (n *nodeBase) removeChildNode(child Node) {
	idx := slices.Index(n.children, child)

	if idx == -1 {
		return
	}

	n.children = slices.Delete(n.children, idx, idx+1)
}

func (n *nodeBase) setParent(parent *Container) {
	if n.parent == parent {
		return
	}

	if n.parent != nil {
		n.parent.removeChildNode(n.self)
	}

	n.parent = parent

	if n.parent != nil {
		n.parent.addChildNode(n.self)
	} else {
		n.detachFromGraph(nil)
	}
}

func (n *nodeBase) attachToGraph(g *Graph) {
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

func (n *nodeBase) detachFromGraph(g *Graph) {
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
