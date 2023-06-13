package psi

import (
	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
	"github.com/zeroflucs-given/generics/collections/stack"
)

type Container struct {
	NodeBase

	node     dst.Node
	comments []string
}

func (c *Container) IsContainer() bool  { return true }
func (c *Container) IsLeaf() bool       { return false }
func (c *Container) Comments() []string { return c.comments }
func (c *Container) Ast() dst.Node      { return c.node }

type Leaf struct {
	NodeBase

	node     dst.Node
	comments []string
}

func (f *Leaf) IsContainer() bool  { return false }
func (f *Leaf) IsLeaf() bool       { return true }
func (f *Leaf) Ast() dst.Node      { return f.node }
func (f *Leaf) Comments() []string { return f.comments }

func convertNode(root dst.Node, sf *SourceFile) (result Node) {
	containerStack := stack.NewStack[Node](16)

	dstutil.Apply(root, func(cursor *dstutil.Cursor) bool {
		node := cursor.Node()

		if node == nil {
			return false
		}

		_, parent := containerStack.Peek()

		wrapped := wrapNode(node)
		wrapped.setParent(parent)

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

func Clone(n Node) Node {
	return wrapNode(cloneTree(n.Ast()))
}

func wrapNode(node dst.Node) Node {
	switch node.(type) {
	case *dst.File:
		return buildContainer(node)
	case *dst.FuncDecl:
		return buildContainer(node)
	case *dst.GenDecl:
		return buildContainer(node)
	case *dst.TypeSpec:
		return buildContainer(node)
	case *dst.ImportSpec:
		return buildContainer(node)
	case *dst.ValueSpec:
		return buildContainer(node)
	default:
		return buildLeaf(node)
	}
}

func buildContainer(node dst.Node) *Container {
	c := &Container{
		node: node,
	}

	c.Initialize(c)

	_, _, dec := dstutil.Decorations(node)

	for _, d := range dec {
		c.comments = append(c.comments, d.Decs...)
	}

	return c
}

func buildLeaf(node dst.Node) *Leaf {
	l := &Leaf{
		node: node,
	}

	l.Initialize(l)

	_, _, dec := dstutil.Decorations(node)

	for _, d := range dec {
		l.comments = append(l.comments, d.Decs...)
	}

	return l
}

func cloneTree(node dst.Node) dst.Node {
	return dst.Clone(node)
}
