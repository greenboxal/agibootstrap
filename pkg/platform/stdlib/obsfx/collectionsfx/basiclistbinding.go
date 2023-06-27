package collectionsfx

import (
	obsfx2 "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

func BindListAny[Dst ModifiableObservableList[T], T any](
	dst Dst,
	src BasicObservableList,
	mapper func(any) T,
) obsfx2.Binding[ObservableList[T]] {
	b := &basicListBinding[T]{
		dst:    dst,
		src:    src,
		mapper: mapper,
	}

	dst.Clear()

	src.AddBasicListListener(b)

	dst.Bind(b)

	b.OnListChangedRaw(&listChangeEvent[T]{
		list: src,
		from: 0,
		to:   src.Len(),
	})

	return b
}

type basicListBinding[T any] struct {
	obsfx2.ObservableValueBase[ObservableList[T]]

	valid  bool
	dst    ModifiableObservableList[T]
	src    BasicObservableList
	mapper func(any) T
}

func (m *basicListBinding[T]) Dependencies() []obsfx2.Observable {
	return []obsfx2.Observable{m.src}
}

func (m *basicListBinding[T]) Invalidate() {
	if m.valid {
		m.valid = false

		m.ObservableValueBase.OnInvalidated(m)
	}
}

func (m *basicListBinding[T]) IsValid() bool {
	return m.valid
}

func (m *basicListBinding[T]) OnListChangedRaw(ev BasicListChangeEvent) {
	for ev.Next() {
		if ev.WasPermutated() {
			m.dst.RemoveCount(ev.From(), ev.To())

			for i := ev.From(); i < ev.To(); i++ {
				v := ev.BasicList().RawGet(i)
				m.dst.InsertAt(ev.From()+i, m.mapper(v))
			}
		} else {
			if ev.WasRemoved() {
				m.dst.RemoveCount(ev.From(), ev.RemovedCount())
			}

			if ev.WasAdded() {
				for i := ev.From(); i < ev.To(); i++ {
					v := ev.BasicList().RawGet(i)
					m.dst.InsertAt(ev.From()+i, m.mapper(v))
				}
			}
		}
	}

	m.valid = true

	m.Invalidate()
}

func (o *basicListBinding[T]) RawValue() any {
	return o.Value()
}

func (m *basicListBinding[T]) Value() ObservableList[T] {
	return m.dst
}

func (m *basicListBinding[T]) Close() {
	m.src.RemoveBasicListListener(m)
}
