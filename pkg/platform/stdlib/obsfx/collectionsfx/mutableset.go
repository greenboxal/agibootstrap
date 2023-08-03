package collectionsfx

type MutableSet[T comparable] struct {
	ObservableSetBase[T]

	keys map[T]struct{}
}

func (s *MutableSet[T]) Has(key T) bool {
	_, ok := s.keys[key]

	return ok
}

func (s *MutableSet[T]) Add(key T) bool {
	_, exists := s.keys[key]

	if exists {
		return false
	}

	s.keys[key] = struct{}{}

	s.FireSetChanged(SetChangeEvent[T]{
		Set:      s,
		Key:      key,
		WasAdded: true,
	})

	return true
}

func (s *MutableSet[T]) Remove(key T) bool {
	_, exists := s.keys[key]

	if !exists {
		return false
	}

	delete(s.keys, key)

	s.FireSetChanged(SetChangeEvent[T]{
		Set:        s,
		Key:        key,
		WasRemoved: true,
	})

	return true
}

func (s *MutableSet[T]) Slice() []T {
	result := make([]T, 0, len(s.keys))

	for k := range s.keys {
		result = append(result, k)
	}

	return result
}

func (s *MutableSet[T]) Len() int {
	return len(s.keys)
}

func (s *MutableSet[T]) Iterator() Iterator[T] {
	return &mutableSetIterator[T]{set: s}
}

func (s *MutableSet[T]) Clear() {
	for k := range s.keys {
		s.Remove(k)
	}
}

type mutableSetIterator[K comparable] struct {
	set   *MutableSet[K]
	index int
	keys  []K
}

func (it *mutableSetIterator[T]) Value() T {
	return it.Item()
}

func (it *mutableSetIterator[T]) Item() T {
	if it.index > len(it.keys) {
		panic("out of range")
	}

	return it.keys[it.index-1]
}

func (it *mutableSetIterator[T]) Next() bool {
	if it.index == 0 && it.keys == nil {
		it.keys = it.set.Slice()
	}

	if it.index >= len(it.keys) {
		it.keys = nil
		return false
	}

	it.index++

	return true
}

func (it *mutableSetIterator[T]) Reset() {
	it.index = 0
	it.keys = nil
}
