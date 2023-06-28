package iterators

import "sort"

type Comparator[T any] func(T, T) int

type sortIterator[T any] struct {
	src        Iterator[T]
	sorted     Iterator[T]
	comparator func(T, T) int
	current    []T
}

func (s *sortIterator[T]) Next() bool {
	if s.sorted == nil {
		result := make([]T, 0)

		for s.src.Next() {
			result = append(result, s.src.Value())
		}

		sort.Slice(result, func(i, j int) bool {
			return s.comparator(result[i], result[j]) < 0
		})

		s.sorted = FromSlice(result)
	}

	return s.sorted.Next()
}

func (s *sortIterator[T]) Value() T {
	return s.sorted.Value()
}

func SortWith[T any](src Iterator[T], comparator Comparator[T]) Iterator[T] {
	return &sortIterator[T]{
		src:        src,
		comparator: comparator,
	}
}
