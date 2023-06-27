package stdlib

type Collection[T any] interface {
	Iterable[T]

	Add(value T) int
	Get(index int) T
	InsertAt(index int, value T)
	Remove(value T) bool
	RemoveAt(index int) T
	IndexOf(value T) int
	Contains(value T) bool
	Length() int
}
