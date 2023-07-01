package psi

import "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"

func NewCursor() Cursor {
	return &cursor{}
}

// Cursor is a stateful tree traversal interface.
type Cursor interface {
	Depth() int

	// Next advances the cursor to the next node.
	Next() bool

	// Current returns the current node.
	Value() Node
	// SetCurrent sets the current node.
	SetCurrent(node Node)
	// SetNext sets the next node.
	SetNext(node Node)
	// Enqueue enqueues the given node iterator.
	Enqueue(it iterators.Iterator[Node])
	// Push pushes the given node iterator.
	Push(it iterators.Iterator[Node])
	// Pop pops the current node iterator.
	Pop() bool
	// PushChildren pushes the children of the current node.
	PushChildren()
	// PushEdges pushes the edges of the current node.
	PushEdges()

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
	depth int

	current  Node
	iterator NodeIterator

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

func (c *cursor) Value() Node   { return c.state.current }
func (c *cursor) Depth() int    { return c.state.depth }
func (c *cursor) WalkChildren() { c.state.walkChildren = true }
func (c *cursor) SkipChildren() { c.state.walkChildren = false }
func (c *cursor) WalkEdges()    { c.state.walkEdges = true }
func (c *cursor) SkipEdges()    { c.state.walkEdges = false }

func (c *cursor) SetCurrent(node Node) {
	c.state.current = node
}

func (c *cursor) SetNext(node Node) {
	it := (&nodeSliceIterator{items: []Node{node}}).Append(c.state.iterator)

	c.state.iterator = it
}

func (c *cursor) Enqueue(it iterators.Iterator[Node]) {
	c.state.iterator = it

	if c.state.iterator == nil {
		c.state.iterator = it
	} else {
		c.state.iterator = &nestedNodeIterator{
			iterators: []NodeIterator{c.state.iterator, it},
		}
	}
}

func (c *cursor) InsertBefore(newNode Node) {
	n := c.Value()
	p := n.Parent()

	if p == nil {
		panic("cannot insert after root node")
	}

	p.InsertChildBefore(n, newNode)
}

func (c *cursor) InsertAfter(newNode Node) {
	n := c.Value()
	p := n.Parent()

	if p == nil {
		panic("cannot insert after root node")
	}

	p.InsertChildAfter(n, newNode)
}

func (c *cursor) Replace(newNode Node) {
	n := c.Value()
	p := n.Parent()

	if p != nil {
		p.ReplaceChildNode(n, newNode)
	}

	c.SetCurrent(newNode)
}

func (c *cursor) Next() bool {
	if !c.state.iterator.Next() {
		return false
	}

	c.state.current = c.state.iterator.Value()

	return true
}

func (c *cursor) EnqueueChildren() {
	c.Enqueue(c.state.current.ChildrenIterator())
}

func (c *cursor) PushEdges() {
	//c.Push(c.state.depth+1, c.state.current.EdgesIterator())
	panic("not implemented")
}

func (c *cursor) Push(it iterators.Iterator[Node]) {
	st := cursorState{
		depth:    c.state.depth + 1,
		iterator: it,

		walkChildren: c.walkChildren,
		walkEdges:    c.walkEdges,
	}

	c.push(st)
}

func (c *cursor) Pop() bool {
	if len(c.stack) == 0 {
		return false
	}

	c.pop()

	return true
}

func (c *cursor) PushChildren() {
	c.Push(c.state.current.ChildrenIterator())
}

func (c *cursor) Walk(n Node, walkFn WalkFunc) (err error) {
	c.Enqueue(&nodeSliceIterator{items: []Node{n}})

	seenMap := map[Node]int{}

	for {
		if c.Next() {
			if c.state.current == nil {
				break
			}

			if seenMap != nil {
				if seenMap[c.state.current] != 0 {
					continue
				}

				seenMap[c.state.current]++
			}

			c.state.walkEdges = c.walkEdges
			c.state.walkChildren = c.walkChildren

			if err := walkFn(c, true); err != nil {
				return err
			}

			if c.state.walkEdges {
				c.PushEdges()
			}

			if c.state.walkChildren {
				c.PushChildren()
			}
		} else {
			if len(c.stack) == 0 {
				break
			}

			c.pop()

			if c.state.current == nil {
				break
			}

			if seenMap != nil {
				if seenMap[c.state.current] != 1 {
					continue
				}

				seenMap[c.state.current] *= -1
			}

			if err := walkFn(c, false); err != nil {
				return err
			}
		}
	}

	return nil
}
