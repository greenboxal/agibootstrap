package psi

// Cursor is a stateful tree traversal interface.
type Cursor interface {
	// Current returns the current node.
	Current() Node
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

	// Replace replaces the current node with the given node.
	Replace(node Node)
	// InsertAfter inserts the given node before the current node.
	InsertAfter(decl Node)
}

type cursorState struct {
	current      Node
	walkChildren bool
	walkEdges    bool
}

type cursor struct {
	walkChildren bool
	stack        []cursorState
	last         cursorState
}

func (c *cursor) push(st cursorState) {
	c.stack = append(c.stack, st)
}

func (c *cursor) pop() cursorState {
	st := c.stack[len(c.stack)-1]
	c.stack = c.stack[:len(c.stack)-1]
	return st
}

func (c *cursor) state() *cursorState {
	return &c.stack[len(c.stack)-1]
}

func (c *cursor) Current() Node {
	return c.state().current
}

func (c *cursor) WalkChildren() {
	c.state().walkChildren = true
}

func (c *cursor) SkipChildren() {
	c.state().walkChildren = false
}

func (c *cursor) WalkEdges() {
	c.state().walkEdges = true
}

func (c *cursor) SkipEdges() {
	c.state().walkEdges = false
}

func (c *cursor) SetCurrent(node Node) {
	c.state().current = node
}

func (c *cursor) InsertBefore(newNode Node) {
	n := c.Current()
	p := n.Parent()

	if p == nil {
		panic("cannot insert after root node")
	}

	p.insertChildNodeBefore(n, newNode)
}

func (c *cursor) InsertAfter(newNode Node) {
	n := c.Current()
	p := n.Parent()

	if p == nil {
		panic("cannot insert after root node")
	}

	p.insertChildNodeAfter(n, newNode)
}

func (c *cursor) Replace(newNode Node) {
	n := c.Current()
	p := n.Parent()

	if p != nil {
		p.replaceChildNode(n, newNode)
	}

	c.SetCurrent(newNode)
}

func (c *cursor) Walk(n Node, walkFn WalkFunc) (err error) {
	c.push(cursorState{current: n, walkChildren: true, walkEdges: false})

	defer func() {
		c.last = c.pop()
	}()

	if err := walkFn(c, true); err != nil {
		return err
	}

	if c.state().walkChildren {
		for _, child := range n.Children() {
			if err := c.Walk(child, walkFn); err != nil {
				return err
			}
		}
	} else {
		return nil
	}

	if err := walkFn(c, false); err != nil {
		return err
	}

	if c.state().walkEdges {
		panic("not implemented")
	}

	return nil
}

// WalkFunc is the type of the function called for each node visited by Walk.
type WalkFunc func(cursor Cursor, entering bool) error

// Walk traverses a PSI Tree in depth-first order.
func Walk(node Node, walkFn WalkFunc) error {
	c := &cursor{}

	return c.Walk(node, walkFn)
}

// Rewrite traverses a PSI Tree in depth-first order and rewrites it.
func Rewrite(node Node, walkFunc WalkFunc) (Node, error) {
	c := &cursor{}

	if err := c.Walk(node, walkFunc); err != nil {
		return nil, err
	}

	return c.last.current, nil
}
