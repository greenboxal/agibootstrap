package psi

import "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"

func GetEdge[T Node](n Node, key TypedEdgeReference[T]) (def T, ok bool) {
	e := n.GetEdge(key)

	if e == nil {
		return def, false
	}

	if v := e.To(); v != nil {
		return v.(T), true
	}

	return def, false
}

func LoadEdge[K TypedEdgeReference[T], T Node](n Node, key K, dst *T) *T {
	e := n.GetEdge(key)

	if e == nil {
		return nil
	}

	if v := e.To(); v != nil {
		*dst = e.To().(T)
	} else {
		return nil
	}

	return dst
}

func GetEdgeOrNil[T Node](n Node, key TypedEdgeReference[T]) (def T) {
	e := n.GetEdge(key)

	if e == nil {
		return def
	}

	if v := e.To(); v != nil {
		return v.(T)
	}

	return def
}

func GetEdgeOrDefault[T Node](n Node, key TypedEdgeReference[T], def T) T {
	e := n.GetEdge(key)

	if e == nil {
		return def
	}

	if v := e.To(); v != nil {
		return v.(T)
	}

	return def
}

func GetEdges[T Node](n Node, kind TypedEdgeKind[T]) []T {
	filtered := iterators.Filter(n.Edges(), func(e Edge) bool {
		return e.Kind() == kind.Kind()
	})

	mapped := iterators.Map(filtered, func(e Edge) T {
		return e.To().(T)
	})

	return iterators.ToSlice(mapped)
}

func UpdateEdges[T Node](n Node, kind TypedEdgeKind[T], values []T) {
	filtered := iterators.Filter(n.Edges(), func(e Edge) bool {
		return e.Kind() == kind.Kind()
	})

	for filtered.Next() {
		n.UnsetEdge(filtered.Value().Key())
	}

	for i, v := range values {
		n.SetEdge(kind.Indexed(int64(i)), v)
	}
}

func UpdateEdge[K TypedEdgeReference[T], T Node](n Node, key K, value T) (isNew bool) {
	e := n.GetEdge(key)

	if e == nil {
		isNew = true
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
