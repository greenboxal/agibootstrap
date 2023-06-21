package psi

import "github.com/pkg/errors"

var ErrAbort = errors.New("abort")

// Cursor is a stateful tree traversal interface.
type Cursor interface {
	// Current returns the current node.
	Node() Node
	// SetCurrent sets the current node.
	SetCurrent(node Node)

	// WalkChildren walks the children of the current node.
	WalkChildren()
	// SkipChildren skips the children of the current node.
	SkipChildren()

	// WalkEdges walks the edges of the current node.
	WalkEdges()
	// SkipEdges skips the edges of the current node.
	SkipEdges()

	// Replace replaces the current node with the given node, modifying the AST.
	// If this operation happens during the enter phase, the children of the new node will be visited.
	// If this operation happens during the leave phase, the children of the new node will NOT be visited.
	Replace(node Node)

	// InsertBefore inserts the given node before the current node, modifying the AST.
	// This node will NOT be visited in the current walk.
	InsertBefore(node Node)

	// InsertAfter inserts the given node before the current node, modifying the AST.
	// This node might be visited in the current walk.
	InsertAfter(node Node)
}

type cursorState struct {
	current Node
	queue   []Node

	walkChildren bool
	walkEdges    bool
}

type cursor struct {
	state cursorState
	stack []cursorState

	walkChildren bool
	walkEdges    bool
}

func (c *cursor) push(st cursorState) {
	c.stack = append(c.stack, c.state)
	c.state = st
}

func (c *cursor) pop() cursorState {
	old := c.state
	c.state = c.stack[len(c.stack)-1]
	c.stack = c.stack[:len(c.stack)-1]
	return old
}

func (c *cursor) Node() Node {
	return c.state.current
}

func (c *cursor) WalkChildren() {
	c.state.walkChildren = true
}

func (c *cursor) SkipChildren() {
	c.state.walkChildren = false
}

func (c *cursor) WalkEdges() {
	c.state.walkEdges = true
}

func (c *cursor) SkipEdges() {
	c.state.walkEdges = false
}

func (c *cursor) Enqueue(node Node) {
	c.state.queue = append(c.state.queue, node)
}

func (c *cursor) SetCurrent(node Node) {
	c.state.current = node
}

func (c *cursor) InsertBefore(newNode Node) {
	n := c.Node()
	p := n.Parent()

	if p == nil {
		panic("cannot insert after root node")
	}

	p.insertChildNodeBefore(n, newNode)
}

func (c *cursor) InsertAfter(newNode Node) {
	n := c.Node()
	p := n.Parent()

	if p == nil {
		panic("cannot insert after root node")
	}

	p.insertChildNodeAfter(n, newNode)
}

func (c *cursor) Replace(newNode Node) {
	n := c.Node()
	p := n.Parent()

	if p != nil {
		p.replaceChildNode(n, newNode)
	}

	c.SetCurrent(newNode)
}

func (c *cursor) Next() bool {
	if len(c.state.queue) > 0 {
		c.state.current = c.state.queue[0]
		c.state.queue = c.state.queue[1:]

		return true
	}

	return false
}

func (c *cursor) Walk(n Node, walkFn WalkFunc) (err error) {
	for {
		if c.Next() {
			if c.state.current == nil {
				break
			}

			if err := walkFn(c, true); err != nil {
				return err
			}

			if c.state.walkEdges {
				panic("not implemented")
			}

			if c.state.walkChildren {
				st := cursorState{queue: n.Children(), walkChildren: c.walkChildren, walkEdges: c.walkEdges}

				c.push(st)
			}
		} else {
			if len(c.stack) == 0 {
				break
			}

			c.pop()

			if c.state.current == nil {
				break
			}

			if err := walkFn(c, false); err != nil {
				return err
			}
		}
	}

	return nil
}

// WalkFunc is the type of the function called for each node visited by Walk.
type WalkFunc func(cursor Cursor, entering bool) error

// Walk traverses a PSI Tree in depth-first order.
func Walk(node Node, walkFn WalkFunc) error {
	c := &cursor{
		walkChildren: true,
	}

	return c.Walk(node, walkFn)
}

// Rewrite traverses a PSI Tree in depth-first order and rewrites it.
func Rewrite(node Node, walkFunc WalkFunc) (Node, error) {
	c := &cursor{
		walkChildren: true,
	}

	if err := c.Walk(node, walkFunc); err != nil && err != ErrAbort {
		return nil, err
	}

	return c.state.current, nil
}

func NewCursor() Cursor {
	return &cursor{}
}
