package iterators

type chunkIterator[T any] struct {
	iter  Iterator[T]
	size  int
	chunk Iterator[T]
}

func (c *chunkIterator[T]) Value() Iterator[T] { return c.chunk }

func (c *chunkIterator[T]) Next() bool {
	var chunk []T

	for c.iter.Next() {
		chunk = append(chunk, c.iter.Value())

		if len(chunk) == c.size {
			break
		}
	}

	if len(chunk) == 0 {
		return false
	}

	c.chunk = &sliceIterator[T]{slice: chunk}

	return true
}

type buffered[T any] struct {
	iter   Iterator[T]
	buffer []T
	first  int
	last   int
}

func (b *buffered[T]) Next() bool {
	missing := len(b.buffer) - (b.last - b.first)

	for i := 0; i < missing; i++ {
		if !b.iter.Next() {
			return false
		}

		b.buffer[b.last] = b.iter.Value()
		b.last++

		if b.last >= len(b.buffer) {
			b.last = 0
		}
	}

	if b.first == b.last {
		return false
	}

	b.first++

	if b.first >= len(b.buffer) {
		b.first = 0
	}

	return true
}

func (b *buffered[T]) Value() (def T) {
	if b.first == b.last {
		return
	}

	return b.buffer[b.first]
}
