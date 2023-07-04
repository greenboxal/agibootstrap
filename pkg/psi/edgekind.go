package psi

type EdgeKind string

func (k EdgeKind) String() string {
	return string(k)
}

func (k EdgeKind) Named(name string) EdgeKey {
	return EdgeKey{Kind: k, Name: name}
}

func (k EdgeKind) Indexed(index int64) EdgeKey {
	return EdgeKey{Kind: k, Index: index}
}

func (k EdgeKind) NamedIndexed(name string, index int64) EdgeKey {
	return EdgeKey{Kind: k, Name: name, Index: index}
}

type TypedEdgeKind[T Node] EdgeKind

func (f TypedEdgeKind[T]) Named(name string) TypedEdgeKey[T] {
	return TypedEdgeKey[T]{Kind: f, Name: name}
}

func (f TypedEdgeKind[T]) Indexed(index int64) TypedEdgeKey[T] {
	return TypedEdgeKey[T]{Kind: f, Index: index}
}

func (f TypedEdgeKind[T]) NamedIndexed(name string, index int64) TypedEdgeKey[T] {
	return TypedEdgeKey[T]{Kind: f, Name: name, Index: index}
}

func (f TypedEdgeKind[T]) Singleton() TypedEdgeKey[T] {
	return TypedEdgeKey[T]{Kind: f}
}

func (f TypedEdgeKind[T]) Kind() EdgeKind { return EdgeKind(f) }

var EdgeKindChild = EdgeKind("child")

func DefineEdgeKind[T Node](name string) TypedEdgeKind[T] {
	return TypedEdgeKind[T](name)
}
