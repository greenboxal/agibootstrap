package psi

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
