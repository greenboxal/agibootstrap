package golang

import (
	"strconv"

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

func (nb *NodeBase[T]) Comments() []string { return nb.comments }
func (nb *NodeBase[T]) Ast() dst.Node      { return nb.node }

func (nb *NodeBase[T]) Initialize(self Node) {
	nb.NodeBase.Init(self, "")
}

func (nb *NodeBase[T]) Update() {
	if nb.IsValid() {
		return
	}

	nb.NodeBase.Update()

	updated := dstutil.Apply(nb.node, func(cursor *dstutil.Cursor) bool {
		n := cursor.Node()

		if n == nil {
			return false
		} else if n == nb.Ast() {
			return true
		}

		k := getEdgeName(cursor.Parent(), cursor.Name(), cursor.Index())
		edge := nb.GetEdge(k)

		if edge != nil {
			to := edge.To().(Node)
			targetNode := to.Ast()

			if targetNode == cursor.Node() {
				cursor.Replace(to.Ast())
			}
		}

		return false
	}, func(cursor *dstutil.Cursor) bool {
		return false
	})

	nb.node = updated.(T)

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

func getEdgeName(parent dst.Node, kind string, index int) psi.EdgeKey {
	name := ""

	if index != -1 {
		name = strconv.FormatInt(int64(index), 10)
	}

	return psi.EdgeKey{
		Kind: psi.EdgeKind(kind),
		Name: name,
	}
}

func AstToPsi(root dst.Node) (result Node) {
	containerStack := stack.NewStack[Node](16)

	dstutil.Apply(root, func(cursor *dstutil.Cursor) bool {
		node := cursor.Node()

		if node == nil {
			return false
		}

		_, parent := containerStack.Peek()

		wrapped := NewNodeFor(node)

		if parent != nil {
			wrapped.SetParent(parent)

			key := getEdgeName(node, cursor.Name(), cursor.Index())

			parent.SetEdge(key, wrapped)
		}

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
