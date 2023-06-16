package golang

import (
	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type ApplyFunc func(*Cursor) bool

type Cursor struct {
	*dstutil.Cursor
	node   psi.Node
	parent psi.Node
}

func (c *Cursor) Element() psi.Node { return c.node }
func (c *Cursor) Parent() psi.Node  { return c.parent }

func Apply(root Node, pre, post ApplyFunc) (result psi.Node) {
	c := &Cursor{}
	refMap := make(map[dst.Node]psi.Node)

	refMap[root.Ast()] = root

	return NewNodeFor(dstutil.Apply(root.Ast(), func(cursor *dstutil.Cursor) bool {
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

		if n.IsContainer() {
			c.parent = n

			for _, child := range n.Children() {
				refMap[child.(Node).Ast()] = child
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

		if c.parent != nil {
			if c.parent == n {
				c.parent = c.parent.Parent()
			}
		}

		if post == nil {
			return true
		}

		return post(c)
	}))
}
