package iterators

type mapIterator[T, U any] struct {
	iter  Iterator[T]
	fn    func(T) U
	value U
}

func (m *mapIterator[T, U]) Next() bool {
	if !m.iter.Next() {
		return false
	}

	m.value = m.fn(m.iter.Value())

	return true
}

func (m *mapIterator[T, U]) Value() U { return m.value }

type flatMapIterator[T, U any] struct {
	src     Iterator[T]
	fn      func(T) Iterator[U]
	current Iterator[U]
	value   U
}

func (f *flatMapIterator[T, U]) Next() bool {
	for {
		if f.current == nil {
			if !f.src.Next() {
				return false
			}

			f.current = f.fn(f.src.Value())
		}

		if !f.current.Next() {
			f.current = nil

			continue
		}

		f.value = f.current.Value()

		return true
	}
}

func (f *flatMapIterator[T, U]) Value() U { return f.value }
