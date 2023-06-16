package psi

type Cursor interface {
	Current() Node
	SetCurrent(node Node)

	WalkChildren()
	SkipChildren()

	WalkEdges()
	SkipEdges()
}

type cursor struct {
	current Node
}

func (c *cursor) SetCurrent(node Node) {

}

type WalkFunc func(cursor Cursor) error

func Walk(node Node, walkFn WalkFunc) error {
	return nil
}
