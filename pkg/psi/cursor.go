package psi

func NewCursor() Cursor {
	return &cursor{}
}

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

func (c *cursor) Enqueue(it NodeIterator) {
	if c.state.iterator == nil {
		c.state.iterator = it
	} else {
		c.state.iterator = &nestedNodeIterator{
			iterators: []NodeIterator{c.state.iterator, it},
		}
	}
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
	if !c.state.iterator.Next() {
		return false
	}

	c.state.current = c.state.iterator.Node()

	return true
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
				panic("not implemented")
			}

			if c.state.walkChildren {
				st := cursorState{
					iterator: c.state.current.ChildrenIterator(),

					walkChildren: c.walkChildren,
					walkEdges:    c.walkEdges,
				}

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
