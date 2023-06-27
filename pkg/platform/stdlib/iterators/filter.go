package iterators

import "reflect"

type filterIterator[T any] struct {
	iter Iterator[T]
	fn   func(T) bool
}

func (f *filterIterator[T]) Value() T { return f.iter.Value() }

func (f *filterIterator[T]) Next() bool {
	for {
		if !f.iter.Next() {
			return false
		}

		if !f.fn(f.iter.Value()) {
			continue
		}

		return true
	}
}

type mapFilterIterator[T, U any] struct {
	iter    Iterator[T]
	typ     reflect.Type
	current U
}

func (m *mapFilterIterator[T, U]) Next() bool {
	for m.iter.Next() {
		v := reflect.ValueOf(m.iter.Value())

		if v.Type().ConvertibleTo(m.typ) {
			m.current = v.Convert(m.typ).Interface().(U)
			return true
		}
	}

	return false
}

func (m *mapFilterIterator[T, U]) Value() U {
	return m.current
}
