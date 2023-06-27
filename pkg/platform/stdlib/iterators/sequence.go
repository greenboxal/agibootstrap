package iterators

type onEachIterator[T any] struct {
	iter Iterator[T]
	fn   func(T)
}

func (o *onEachIterator[T]) Next() bool {
	if !o.iter.Next() {
		return false
	}

	o.fn(o.iter.Value())

	return true
}

func (o *onEachIterator[T]) Value() T {
	return o.iter.Value()
}

type scanIterator[T, U any] struct {
	iter    Iterator[T]
	current U
	fn      func(U, T) U
}

func (s *scanIterator[T, U]) Next() bool {
	if !s.iter.Next() {
		return false
	}

	s.current = s.fn(s.current, s.iter.Value())

	return true
}

func (s *scanIterator[T, U]) Value() U {
	return s.current
}

type concatIterator[T any] struct {
	iterators []Iterator[T]
	current   Iterator[T]
}

func (c *concatIterator[T]) Next() bool {
	for {
		if c.current == nil {
			if len(c.iterators) == 0 {
				return false
			}

			c.current = c.iterators[0]
			c.iterators = c.iterators[1:]
		}

		if !c.current.Next() {
			c.current = nil

			continue
		}

		return true
	}
}

func (c *concatIterator[T]) Value() (def T) {
	if c.current == nil {
		return
	}

	return c.current.Value()
}
