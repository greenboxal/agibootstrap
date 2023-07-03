package promptml

func Container(children ...AttachableNodeLike) Parent {
	c := NewContainer()

	for _, child := range children {
		child.SetParent(c)
	}

	return c
}

func MakeFixed[T Node](t T) T {
	t.SetResizable(false)

	return t
}
