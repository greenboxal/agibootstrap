package golang

import (
	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
	"github.com/zeroflucs-given/generics/collections/stack"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Node interface {
	psi.Node

	Ast() dst.Node
	Initialize(self Node)
}

type NodeBase[T dst.Node] struct {
	psi.NodeBase

	node     T
	comments []string
}

func (nb *NodeBase[T]) IsContainer() bool { return true }
func (nb *NodeBase[T]) IsLeaf() bool      { return false }

func (nb *NodeBase[T]) Comments() []string { return nb.comments }
func (nb *NodeBase[T]) Ast() dst.Node      { return nb.node }

func (nb *NodeBase[T]) Initialize(self Node) {
	nb.NodeBase.Init(self, "")
}

type Container struct {
	NodeBase[dst.Node]
}

func (c *Container) IsContainer() bool { return true }
func (c *Container) IsLeaf() bool      { return false }

type Leaf struct {
	NodeBase[dst.Node]
}

func (f *Leaf) IsContainer() bool { return false }
func (f *Leaf) IsLeaf() bool      { return true }

func AstToPsi(root dst.Node) (result Node) {
	containerStack := stack.NewStack[Node](16)

	dstutil.Apply(root, func(cursor *dstutil.Cursor) bool {
		node := cursor.Node()

		if node == nil {
			return false
		}

		_, parent := containerStack.Peek()

		wrapped := NewNodeFor(node)
		wrapped.SetParent(parent)

		if wrapped.IsContainer() {
			if err := containerStack.Push(wrapped.(*Container)); err != nil {
				panic(err)
			}
		}

		return true
	}, func(cursor *dstutil.Cursor) bool {
		node := cursor.Node()
		hasParent, parent := containerStack.Peek()

		if hasParent && parent.Ast() == node {
			_, result = containerStack.Pop()
		}

		return true
	})

	return
}

func NewNodeFor(node dst.Node) Node {
	switch node.(type) {
	case *dst.File:
		return NewContainer(node)
	case *dst.FuncDecl:
		return NewContainer(node)
	case *dst.GenDecl:
		return NewContainer(node)
	case *dst.TypeSpec:
		return NewContainer(node)
	case *dst.ImportSpec:
		return NewContainer(node)
	case *dst.ValueSpec:
		return NewContainer(node)
	default:
		return NewLeaf(node)
	}
}

func NewContainer(node dst.Node) *Container {
	c := &Container{}

	c.node = node

	c.Initialize(c)

	_, _, dec := dstutil.Decorations(node)

	for _, d := range dec {
		c.comments = append(c.comments, d.Decs...)
	}

	return c
}

func NewLeaf(node dst.Node) *Leaf {
	l := &Leaf{}

	l.node = node

	l.Initialize(l)

	_, _, dec := dstutil.Decorations(node)

	for _, d := range dec {
		l.comments = append(l.comments, d.Decs...)
	}

	return l
}
