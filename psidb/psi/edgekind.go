package psi

type EdgeKind string

func (k EdgeKind) Type() EdgeType {
	return LookupEdgeType(k)
}

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

func (k EdgeKind) Kind() EdgeKind { return k }

type TypedEdgeKind[T Node] EdgeKind

func (k TypedEdgeKind[T]) Type() EdgeType {
	return LookupEdgeType(k.Kind())
}

func (k TypedEdgeKind[T]) Named(name string) TypedEdgeKey[T] {
	return TypedEdgeKey[T]{Kind: k, Name: name}
}

func (k TypedEdgeKind[T]) Indexed(index int64) TypedEdgeKey[T] {
	return TypedEdgeKey[T]{Kind: k, Index: index}
}

func (k TypedEdgeKind[T]) NamedIndexed(name string, index int64) TypedEdgeKey[T] {
	return TypedEdgeKey[T]{Kind: k, Name: name, Index: index}
}

func (k TypedEdgeKind[T]) Singleton() TypedEdgeKey[T] {
	return TypedEdgeKey[T]{Kind: k}
}

func (k TypedEdgeKind[T]) Kind() EdgeKind { return EdgeKind(k) }

var EdgeKindRoot = DefineEdgeType[Node]("root").Kind()
var EdgeKindChild = DefineEdgeType[Node]("child").Kind()
