package collectionsfx

type Iterable[E any] interface {
	Iterator() Iterator[E]
}

type Iterator[E any] interface {
	Value() E
	Item() E
	Next() bool
	Reset()
}

type ListIterator[E any] interface {
	Iterator[E]

	Index() int
}
