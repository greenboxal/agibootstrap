package iterators

import (
	"context"
	"reflect"
)

func Concat[IT Iterator[T], T any](iterators ...IT) Iterator[T] {
	iters := make([]Iterator[T], len(iterators))

	for i, iter := range iterators {
		iters[i] = iter
	}

	return &concatIterator[T]{iterators: iters}
}

func Filter[IT Iterator[T], T any](iter IT, fn func(T) bool) Iterator[T] {
	return &filterIterator[T]{iter: iter, fn: fn}
}

func FilterIsInstance[T, U any](iter Iterator[T]) Iterator[U] {
	return &mapFilterIterator[T, U]{iter: iter, typ: reflect.TypeOf((*U)(nil)).Elem()}
}

func Map[IT Iterator[T], T, U any](iter IT, fn func(T) U) Iterator[U] {
	return &mapIterator[T, U]{iter: iter, fn: fn}
}

func FlatMap[IT Iterator[T], T, U any](iter IT, fn func(T) Iterator[U]) Iterator[U] {
	return &flatMapIterator[T, U]{src: iter, fn: fn}
}

func Count[IT Iterator[T], T any](iter IT) int {
	var count int

	for iter.Next() {
		count++
	}

	return count
}

func Buffer[T any](iter Iterator[T], size int) Iterator[Iterator[T]] {
	return &chunkIterator[T]{iter: iter, size: size}
}

func Buffered[T any](iter Iterator[T], size int) Iterator[T] {
	return &buffered[T]{iter: iter, buffer: make([]T, size)}
}

func OnEach[T any](iter Iterator[T], fn func(T)) Iterator[T] {
	return &onEachIterator[T]{iter: iter, fn: fn}
}

func Fold[IT Iterator[T], T any, U any](iter IT, initial U, fn func(U, T) U) U {
	for iter.Next() {
		initial = fn(initial, iter.Value())
	}

	return initial
}

func Reduce[IT Iterator[T], T any, U any](iter IT, reducer func(Iterator[T]) U) U {
	return reducer(iter)
}

func Scan[IT Iterator[T], T any, U any](iter IT, initial U, fn func(U, T) U) Iterator[U] {
	return &scanIterator[T, U]{iter: iter, current: initial, fn: fn}
}

func ForEach[IT Iterator[T], T any](iter IT, fn func(T)) {
	for iter.Next() {
		fn(iter.Value())
	}
}

func ForEachContext[IT Iterator[T], T any](ctx context.Context, iter IT, fn func(context.Context, T)) {
	for iter.Next() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		fn(ctx, iter.Value())

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}
