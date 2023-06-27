package collectionsfx

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type mapBinding[K comparable, V, T any] struct {
	obsfx.ObservableValueBase[ObservableMap[K, T]]

	dst    ModifiableObservableMap[K, T]
	src    ObservableMap[K, V]
	mapper func(K, V) T
}

func BindMap[Src ObservableMap[K, V], Dst ModifiableObservableMap[K, T], K comparable, V, T any](
	dst Dst,
	src Src,
	mapper func(K, V) T,
) obsfx.Binding[ObservableMap[K, T]] {
	b := &mapBinding[K, V, T]{
		dst:    dst,
		src:    src,
		mapper: mapper,
	}

	src.AddMapListener(b)

	for k, v := range src.Map() {
		b.OnMapChanged(MapChangeEvent[K, V]{
			Key:        k,
			ValueAdded: v,
			WasAdded:   true,
		})
	}

	return b
}

func (m *mapBinding[K, V, T]) Dependencies() []obsfx.Observable {
	return []obsfx.Observable{m.src}
}

func (m *mapBinding[K, V, T]) Invalidate() {
}

func (m *mapBinding[K, V, T]) IsValid() bool {
	return true
}

func (m *mapBinding[K, V, T]) OnMapChanged(ev MapChangeEvent[K, V]) {
	if ev.WasAdded {
		m.dst.Set(ev.Key, m.mapper(ev.Key, ev.ValueAdded))
	} else if ev.WasRemoved {
		m.dst.Remove(ev.Key)
	}

	m.ObservableValueBase.OnInvalidated(m)
}

func (o *mapBinding[K, V, T]) RawValue() any {
	return o.Value()
}

func (m *mapBinding[K, V, T]) Value() ObservableMap[K, T] {
	return m.dst
}

func (m *mapBinding[K, V, T]) Close() {
	m.src.RemoveMapListener(m)
}
