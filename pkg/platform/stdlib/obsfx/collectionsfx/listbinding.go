package collectionsfx

import (
	obsfx2 "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type ListBindingInterface[T any] interface {
	Set(index int, value T)
	Swap(i, j int)
	InsertAll(index int, value ...T)
	RemoveCount(index int, count int)
}

func ObserveList[Src ObservableList[T], T any](src Src, cb func(ev ListChangeEvent[T])) {
	src.AddListListener(OnListChangedFunc[T](cb))

	cb(&listChangeEvent[T]{
		list: src,
		from: 0,
		to:   src.Len(),
	})
}

func BindListInterface[Dst ListBindingInterface[T], Src ObservableList[V], V, T any](dst Dst, src Src, mapper func(V) T) obsfx2.Binding[ObservableList[T]] {
	temp := &MutableSlice[T]{}
	binding := BindList(temp, src, mapper)

	ObserveList(temp, func(ev ListChangeEvent[T]) {
		for ev.Next() {
			if ev.WasPermutated() {
				dst.RemoveCount(ev.From(), ev.To())

				for i, v := range ev.List().SubSlice(ev.From(), ev.To()) {
					dst.InsertAll(ev.From()+i, v)
				}
			} else {
				if ev.WasRemoved() {
					dst.RemoveCount(ev.From(), ev.RemovedCount())
				}

				if ev.WasAdded() {
					for i, v := range ev.AddedSlice() {
						dst.InsertAll(ev.From()+i, v)
					}
				}
			}
		}
	})

	return binding
}

func BindList[Dst ModifiableObservableList[T], Src ObservableList[V], V, T any](
	dst Dst,
	src Src,
	mapper func(V) T,
) obsfx2.Binding[ObservableList[T]] {
	b := &listBinding[V, T]{
		dst:    dst,
		src:    src,
		mapper: mapper,
	}

	dst.Clear()

	src.AddListListener(b)

	dst.Bind(b)

	b.OnListChanged(&listChangeEvent[V]{
		list: src,
		from: 0,
		to:   src.Len(),
	})

	return b
}

type listBinding[V, T any] struct {
	obsfx2.ObservableValueBase[ObservableList[T]]

	valid  bool
	dst    ModifiableObservableList[T]
	src    ObservableList[V]
	mapper func(V) T
}

func (m *listBinding[V, T]) Dependencies() []obsfx2.Observable {
	return []obsfx2.Observable{m.src}
}

func (m *listBinding[V, T]) Invalidate() {
	if m.valid {
		m.valid = false

		m.ObservableValueBase.OnInvalidated(m)
	}
}

func (m *listBinding[V, T]) IsValid() bool {
	return m.valid
}

func (m *listBinding[V, T]) OnListChanged(ev ListChangeEvent[V]) {
	for ev.Next() {
		if ev.WasPermutated() {
			m.dst.RemoveCount(ev.From(), ev.To())

			for i, v := range ev.List().SubSlice(ev.From(), ev.To()) {
				m.dst.InsertAt(ev.From()+i, m.mapper(v))
			}
		} else {
			if ev.WasRemoved() {
				m.dst.RemoveCount(ev.From(), ev.RemovedCount())
			}

			if ev.WasAdded() {
				for i, v := range ev.AddedSlice() {
					m.dst.InsertAt(ev.From()+i, m.mapper(v))
				}
			}
		}
	}

	m.valid = true

	m.Invalidate()
}

func (o *listBinding[V, T]) RawValue() any {
	return o.Value()
}

func (m *listBinding[V, T]) Value() ObservableList[T] {
	return m.dst
}

func (m *listBinding[V, T]) Close() {
	m.src.RemoveListListener(m)
}

type listMapBinding[K comparable, V, T any] struct {
	obsfx2.ObservableValueBase[ObservableList[T]]

	valid   bool
	dst     ModifiableObservableList[T]
	src     ObservableMap[K, V]
	mapper  func(K, V) T
	indexes map[K]int
}

func BindListFromMap[Dst ModifiableObservableList[T], Src ObservableMap[K, V], K comparable, V, T any](dst Dst, src Src, mapper func(K, V) T) obsfx2.Binding[ObservableList[T]] {
	b := &listMapBinding[K, V, T]{
		dst:     dst,
		src:     src,
		mapper:  mapper,
		indexes: map[K]int{},
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

func (m *listMapBinding[K, V, T]) Dependencies() []obsfx2.Observable {
	return []obsfx2.Observable{m.src}
}

func (m *listMapBinding[K, V, T]) Invalidate() {
	if m.valid {
		m.valid = false

		m.ObservableValueBase.OnInvalidated(m)
	}
}

func (m *listMapBinding[K, V, T]) IsValid() bool {
	return m.valid
}

func (m *listMapBinding[K, V, T]) OnMapChanged(ev MapChangeEvent[K, V]) {
	index, hasIndex := m.indexes[ev.Key]

	if ev.WasAdded {
		v := m.mapper(ev.Key, ev.ValueAdded)

		if hasIndex {
			m.dst.Set(index, v)
		} else {
			m.indexes[ev.Key] = m.dst.Add(v)
		}
	} else if ev.WasRemoved {
		if hasIndex {
			delete(m.indexes, ev.Key)

			m.dst.RemoveAt(index)
		}
	}

	m.valid = true

	m.Invalidate()
}

func (o *listMapBinding[K, V, T]) RawValue() any {
	return o.Value()
}

func (m *listMapBinding[K, V, T]) Value() ObservableList[T] {
	return m.dst
}

func (m *listMapBinding[K, V, T]) Close() {
	m.src.RemoveMapListener(m)
}
