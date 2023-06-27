package collectionsfx

type BasicListChangeEvent interface {
	BasicList() BasicObservableList

	From() int
	To() int

	AddedCount() int
	RemovedCount() int

	Permutations() []int
	GetPermutation(i int) int

	WasAdded() bool
	WasRemoved() bool
	WasReplaced() bool
	WasUpdated() bool
	WasPermutated() bool

	Next() bool
}

type ListChangeEvent[T any] interface {
	BasicListChangeEvent

	List() ObservableList[T]

	RemovedSlice() []T
	AddedSlice() []T
}

type listChangeIterator[T any] struct {
	ListChangeEvent[T]

	events []*listChangeEvent[T]
	index  int
}

func (l *listChangeIterator[T]) Next() bool {
	var next ListChangeEvent[T]

	for next == nil {
		if l.index >= len(l.events) {
			return false
		}

		next = l.events[l.index]
		l.index++
	}

	l.ListChangeEvent = next

	return true
}

type listChangeEvent[T any] struct {
	list         BasicObservableList
	removedSlice []T
	from         int
	to           int
	updated      bool
	consumed     bool
	perm         []int
}

func (l *listChangeEvent[T]) Next() bool {
	if l.consumed {
		return false
	}

	l.consumed = true

	return true
}

func (l *listChangeEvent[T]) BasicList() BasicObservableList {
	return l.list
}

func (l *listChangeEvent[T]) List() ObservableList[T] {
	return l.list.(ObservableList[T])
}

func (l *listChangeEvent[T]) Permutations() []int {
	return l.perm
}

func (l *listChangeEvent[T]) GetPermutation(i int) int {
	return l.perm[i-l.From()]
}

func (l *listChangeEvent[T]) RemovedSlice() []T {
	return l.removedSlice
}

func (l *listChangeEvent[T]) AddedSlice() []T {
	if !l.WasAdded() {
		return nil
	}

	return l.List().SubSlice(l.from, l.to)
}

func (l *listChangeEvent[T]) WasPermutated() bool {
	return len(l.perm) > 0
}

func (l *listChangeEvent[T]) WasUpdated() bool {
	return l.updated
}

func (l *listChangeEvent[T]) WasAdded() bool {
	return !l.WasPermutated() && !l.WasUpdated() && l.From() < l.To()
}

func (l *listChangeEvent[T]) WasRemoved() bool {
	return len(l.removedSlice) > 0
}

func (l *listChangeEvent[T]) WasReplaced() bool {
	return l.WasAdded() && l.WasRemoved()
}

func (l *listChangeEvent[T]) AddedCount() int {
	if !l.WasAdded() {
		return 0
	}

	return l.to - l.from
}

func (l *listChangeEvent[T]) RemovedCount() int {
	return len(l.removedSlice)
}

func (l *listChangeEvent[T]) From() int {
	return l.from
}

func (l *listChangeEvent[T]) To() int {
	return l.to
}
