package promptml

func Container(children ...Node) Parent {
	c := NewContainer()

	for _, child := range children {
		c.AddChildNode(child)
	}

	return c
}

func Fixed[T Node](t T) T {
	t.SetResizable(false)

	return t
}
