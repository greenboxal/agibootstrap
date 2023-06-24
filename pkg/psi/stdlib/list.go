package stdlib

import (
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type ICollection[T any] interface {
	Add(value T) int
	Get(index int) T
	InsertAt(index int, value T)
	Remove(value T) bool
	RemoveAt(index int) T
	IndexOf(value T) int
	Contains(value T) bool
	Length() int

	Iterator() Iterator[T]
}

type Iterator[T any] interface {
	Next() bool
	Value() T
}

func (c *NodeCollection[T]) PsiNodeName() string { return c.name }

type NodeCollection[T psi.Node] struct {
	psi.NodeBase

	name string
}

func (c *NodeCollection[T]) Add(value T) int {
	index := c.Length()

	c.InsertAt(index, value)

	return index
}

func (c *NodeCollection[T]) Get(index int) T {
	return c.Children()[index].(T)
}

func (c *NodeCollection[T]) InsertAt(index int, value T) {
	c.InsertChildrenAt(index, value)
}

func (c *NodeCollection[T]) Remove(value T) bool {
	index := c.IndexOf(value)

	if index == -1 {
		return false
	}

	c.RemoveAt(index)

	return true
}

func (c *NodeCollection[T]) RemoveAt(index int) T {
	child := c.Children()[index]

	c.RemoveChildNode(child)

	return child.(T)
}

func (c *NodeCollection[T]) IndexOf(value T) int {
	return c.IndexOfChild(value)
}

func (c *NodeCollection[T]) Contains(value T) bool {
	return c.IndexOf(value) != -1
}

func (c *NodeCollection[T]) Length() int {
	return len(c.Children())
}

func (c *NodeCollection[T]) Iterator() Iterator[T] {
	return NewDirectChildrenListIterator(c)
}

type directChildrenListIterator[T psi.Node] struct {
	c       *NodeCollection[T]
	current T
	index   int
}

func (d *directChildrenListIterator[T]) Next() bool {
	if d.index >= d.c.Length() {
		return false
	}

	d.current = d.c.Get(d.index)
	d.index++

	return true
}

func (d *directChildrenListIterator[T]) Value() T {
	return d.current
}

func NewDirectChildrenListIterator[T psi.Node](c *NodeCollection[T]) Iterator[T] {
	return &directChildrenListIterator[T]{
		c: c,
	}
}

func NewNodeCollection[T psi.Node](name string) *NodeCollection[T] {
	col := &NodeCollection[T]{
		name: name,
	}

	col.Init(col, "")

	return col
}