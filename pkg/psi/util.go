package psi

func GetEdge[T Node](n Node, key TypedEdgeReference[T]) (def T, ok bool) {
	e := n.GetEdge(key)

	if e == nil {
		return def, false
	}

	v := e.To().(T)

	return v, true
}

func LoadEdge[K TypedEdgeReference[T], T Node](n Node, key K, dst *T) *T {
	e := n.GetEdge(key)

	if e == nil {
		return nil
	}

	*dst = e.To().(T)

	return dst
}

func GetEdgeOrNil[T Node](n Node, key TypedEdgeReference[T]) (def T) {
	e := n.GetEdge(key)

	if e == nil {
		return def
	}

	v := e.To().(T)

	return v
}

func GetEdgeOrDefault[T Node](n Node, key TypedEdgeReference[T], def T) T {
	e := n.GetEdge(key)

	if e == nil {
		return def
	}

	v := e.To().(T)

	return v
}

func UpdateEdge[K TypedEdgeReference[T], T Node](n Node, key K, value T) (previous T, isNew bool) {
	e := n.GetEdge(key)

	if e == nil {
		isNew = true
	} else {
		previous = e.To().(T)
	}

	n.SetEdge(key, value)

	return
}

func GetOrCreateEdge[K TypedEdgeReference[T], T Node](n Node, key K, fn func() T) T {
	e := n.GetEdge(key)

	if e != nil {
		return e.To().(T)
	}

	t := fn()

	UpdateEdge(n, key, t)

	return t
}

func GetAttribute[T any](n Node, key TypedAttributeKey[T]) (def T, ok bool) {
	e, ok := n.GetAttribute(key.Name())

	if !ok {
		return def, ok
	}

	v := e.(T)

	return v, true
}

func SetAttribute[T any](n Node, key TypedAttributeKey[T], value T) {
	n.SetAttribute(key.Name(), value)
}

func GetOrCreateAttribute[T any](n Node, key TypedAttributeKey[T], fn func(attributeKey TypedAttributeKey[T]) T) T {
	v, ok := GetAttribute(n, key)

	if ok {
		return v
	}

	v = fn(key)

	SetAttribute(n, key, v)

	return v
}

func GetOrCreateAttributeWithDefault[T any](n Node, key TypedAttributeKey[T], def T) T {
	return GetOrCreateAttribute(n, key, func(attributeKey TypedAttributeKey[T]) T {
		return def
	})
}
