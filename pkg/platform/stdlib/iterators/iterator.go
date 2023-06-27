package iterators

import (
	"github.com/go-errors/errors"
)

var ErrStopIteration = errors.New("stop iteration")

type Iterable[T any] interface {
	Iterator() Iterator[T]
}

type Iterator[T any] interface {
	Next() bool
	Value() T
}

type ListIterator[T any] interface {
	Iterator[T]

	Index() int
	Previous() bool
}

type MutableListIterator[T any] interface {
	ListIterator[T]

	InsertBefore(v T)
	InsertAfter(v T)
	Set(v T)
	Remove()
}

type Reducer[T any, U any] interface {
	Reduce(Iterator[T]) U
}

func NewIterator[T any](fn func() (T, bool)) Iterator[T] {
	return &funcIterator[T]{fn: fn}
}

type funcIterator[T any] struct {
	fn      func() (T, bool)
	current T
}

func (f *funcIterator[T]) Next() bool {
	value, ok := f.fn()

	if !ok {
		return false
	}

	f.current = value

	return true
}

func (f *funcIterator[T]) Value() T {
	return f.current
}
