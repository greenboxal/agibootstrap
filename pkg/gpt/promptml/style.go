package promptml

type StyleOpt[T Node] func(v T)

func Styled[I Node, T Node](n T, opts ...StyleOpt[I]) T {
	for _, opt := range opts {
		opt(any(n).(I))
	}

	return n
}

func Fixed() StyleOpt[Node] {
	return func(v Node) {
		v.SetResizable(false)
	}
}
