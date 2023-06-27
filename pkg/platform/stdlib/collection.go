package stdlib

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
)

type Collection[T any] interface {
	iterators.Iterable[T]

	Add(value T) int
	Get(index int) T
	InsertAt(index int, value T)
	Remove(value T) bool
	RemoveAt(index int) T
	IndexOf(value T) int
	Contains(value T) bool
	Len() int
}
