package iterators

func FromSlice[T any](slice []T) Iterator[T] {
	return &sliceIterator[T]{slice: slice}
}

func ToSlice[T any](iter Iterator[T]) []T {
	var result []T

	for iter.Next() {
		result = append(result, iter.Value())
	}

	return result
}

func ToSliceWithCapacity[IT Iterator[T], T any](iter IT, capacity int) []T {
	result := make([]T, 0, capacity)

	for iter.Next() {
		result = append(result, iter.Value())
	}

	return result
}

type sliceIterator[T any] struct {
	slice   []T
	index   int
	current T
}

func (s *sliceIterator[T]) Index() int     { return s.index }
func (s *sliceIterator[T]) Value() (def T) { return s.current }

func (s *sliceIterator[T]) Next() bool {
	if len(s.slice) == 0 {
		return false
	}

	if s.index >= len(s.slice) {
		return false
	}

	s.current = s.slice[s.index]
	s.index++

	return true
}

func (s *sliceIterator[T]) Previous() bool {
	if len(s.slice) == 0 {
		return false
	}

	if s.index <= 0 {
		return false
	}

	s.index--
	s.current = s.slice[s.index]

	return true
}

func (s *sliceIterator[T]) Reset() {
	var empty T

	s.index = 0
	s.current = empty
}
