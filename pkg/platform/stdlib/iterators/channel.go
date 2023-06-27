package iterators

import "context"

func FromChannel[T any](ch <-chan T) Iterator[T] {
	return &channelIterator[T]{ch: ch}
}

func AsChannel[IT Iterator[T], T any](iter IT) <-chan T {
	ch := make(chan T)

	go func() {
		defer close(ch)

		for iter.Next() {
			ch <- iter.Value()
		}
	}()

	return ch
}

func AsBufferedChannel[IT Iterator[T], T any](iter IT, size int) <-chan T {
	ch := make(chan T, size)

	go func() {
		defer close(ch)

		for iter.Next() {
			ch <- iter.Value()
		}
	}()

	return ch
}

func PumpToChannel[IT Iterator[T], T any](ctx context.Context, iter IT, ch chan<- T) {
	for iter.Next() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ch <- iter.Value()

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

type channelIterator[T any] struct {
	ch      <-chan T
	current T
}

func (c *channelIterator[T]) Value() T { return c.current }

func (c *channelIterator[T]) Next() bool {
	var ok bool

	c.current, ok = <-c.ch

	return ok
}
