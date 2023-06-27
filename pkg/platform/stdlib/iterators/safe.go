package iterators

import "context"

type Result[T any] struct {
	Value T
	Err   error
}

func Cancellable[T any](ctx context.Context, iter Iterator[T]) Iterator[T] {
	return &cancellableIterator[T]{ctx: ctx, iter: iter}
}

func Safe[IT Iterator[T], T any](iter IT) Iterator[Result[T]] {
	return &safeIterator[T]{iter: iter}
}

type safeIterator[T any] struct {
	iter    Iterator[T]
	current Result[T]
	empty   T
}

func (s *safeIterator[T]) Next() bool {
	defer func() {
		if err := recover(); err != nil {
			if err == ErrStopIteration {
				s.current.Value = s.empty
				s.current.Err = nil
				return
			}

			if err, ok := err.(error); ok {
				s.current.Err = err
			} else {
				s.current.Err = errors.Wrap(err, 1)
			}
		}
	}()

	s.current.Value = s.empty

	if !s.iter.Next() {
		return false
	}

	s.current.Value = s.iter.Value()

	return true
}

func (s *safeIterator[T]) Value() Result[T] {
	return s.current
}

type cancellableIterator[T any] struct {
	ctx     context.Context
	iter    Iterator[T]
	current T
}

func (c *cancellableIterator[T]) Next() bool {
	defer func() {
		if err := recover(); err != nil {
			if err == ErrStopIteration {
				return
			}

			panic(err)
		}
	}()

	if !c.iter.Next() {
		return false
	}

	select {
	case <-c.ctx.Done():
		return false
	default:
	}

	c.current = c.iter.Value()

	return true
}

func (c *cancellableIterator[T]) Value() T {
	return c.current
}
