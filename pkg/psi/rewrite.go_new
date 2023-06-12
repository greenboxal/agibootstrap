package psi

import (
	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
)

type ApplyFunc func(*Cursor) bool

type Cursor struct {
	*dstutil.Cursor
	node      Node
	container *Container
}

func (c *Cursor) Element() Node         { return c.node }
func (c *Cursor) Container() *Container { return c.container }

func Apply(root Node, pre, post ApplyFunc) (result Node) {
	c := &Cursor{}
	refMap := make(map[dst.Node]Node)

	refMap[root.Node()] = root

	return wrapNode(dstutil.Apply(root.Node(), func(cursor *dstutil.Cursor) bool {
		node := cursor.Node()

		if node == nil {
			return false
		}

		n := refMap[node]

		if n == nil {
			return false
		}

		c.Cursor = cursor
		c.node = n

		if n, ok := n.(*Container); ok {
			c.container = n

			for _, child := range n.Children() {
				refMap[child.Node()] = child
			}
		}

		if pre == nil {
			return true
		}

		return pre(c)
	}, func(cursor *dstutil.Cursor) bool {
		node := cursor.Node()

		if node == nil {
			return true
		}

		n := refMap[node]

		if n == nil {
			return false
		}

		c.Cursor = cursor
		c.node = n

		if c.container != nil {
			if c.container == n {
				c.container = c.container.Parent()
			}
		}

		if post == nil {
			return true
		}

		return post(c)
	}))
}
