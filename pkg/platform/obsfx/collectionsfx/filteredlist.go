package collectionsfx

import (
	"golang.org/x/exp/slices"
)

func NewFilteredList[T any](src ObservableList[T], predicate Predicate[T]) *FilteredList[T] {
	fl := &FilteredList[T]{
		src:       src,
		predicate: predicate,
	}

	src.AddListListener(fl)

	return fl
}

type Predicate[T any] func(a T) bool

type FilteredList[T any] struct {
	ObservableListBase[T]

	filtered   []int
	size       int
	src        ObservableList[T]
	predicate  Predicate[T]
	comparator Comparator[T]
	builder    ListChangeBuilder[T]
}

func (fl *FilteredList[T]) Source() ObservableList[T] {
	return fl.src
}

func (fl *FilteredList[T]) GetSourceIndex(index int) int {
	return fl.filtered[index]
}

func (fl *FilteredList[T]) GetViewIndex(index int) int {
	return slices.Index(fl.filtered, index)
}

func (fl *FilteredList[T]) RawGet(index int) any {
	return fl.Get(index)
}

func (fl *FilteredList[T]) Get(index int) T {
	if index >= fl.size {
		panic("out of bounds")
	}

	return fl.src.Get(fl.filtered[index])
}

func (fl *FilteredList[T]) SubSlice(from, to int) []T {
	result := make([]T, from-to)

	for i := range result {
		result[i] = fl.Get(from + i)
	}

	return result
}

func (fl *FilteredList[T]) Slice() []T {
	return fl.SubSlice(0, fl.Len())
}

func (fl *FilteredList[T]) Len() int {
	return fl.size
}

func (fl *FilteredList[T]) Contains(value T) bool {
	return fl.IndexOf(value) >= 0
}

func (fl *FilteredList[T]) IndexOf(value T) int {
	index := fl.src.IndexOf(value)

	if index == -1 {
		return -1
	}

	return fl.GetViewIndex(index)
}

func (fl *FilteredList[T]) Iterator() Iterator[T] {
	return &listIterator[T]{l: fl}
}

func (fl *FilteredList[T]) getBuilder() *ListChangeBuilder[T] {
	fl.builder.ObservableListBase = &fl.ObservableListBase
	fl.builder.List = fl

	return &fl.builder
}

func (fl *FilteredList[T]) OnListChanged(ev ListChangeEvent[T]) {
	b := fl.getBuilder()
	defer b.End()

	for ev.Next() {
		if ev.WasPermutated() {
			fl.onPermutated(ev)
		} else if ev.WasUpdated() {
			fl.onUpdated(ev)
		} else {
			fl.onAddRemove(ev)
		}
	}
}

func (fl *FilteredList[T]) Close() {
	fl.src.RemoveListListener(fl)
}

func (fl *FilteredList[T]) updateIndexes(from, delta int) {
	for i := from; i < fl.size; i++ {
		fl.filtered[i] += delta
	}
}

func (fl *FilteredList[T]) ensureSize(size int) {
	if len(fl.filtered) < size {
		filtered := make([]int, size*3/2+1)
		copy(filtered, fl.filtered)
		fl.filtered = filtered
	}
}

func (fl *FilteredList[T]) findPosition(index int) int {
	if len(fl.filtered) == 0 {
		return 0
	}

	if index == 0 {
		return 0
	}

	index, ok := slices.BinarySearch(fl.filtered, index)

	if !ok {
		return ^index
	}

	return index
}

func (fl *FilteredList[T]) refilter() {
	fl.ensureSize(fl.src.Len())

	for i := 0; i < fl.src.Len(); i++ {
		val := fl.src.Get(i)

		if fl.predicate(val) {
			fl.filtered[fl.size] = i
			fl.size++
		}
	}
}

func (fl *FilteredList[T]) onPermutated(ev ListChangeEvent[T]) {
	b := fl.getBuilder()
	defer b.End()

	from := fl.findPosition(ev.From())
	to := fl.findPosition(ev.To())

	if to > from {
		for i := from; i < to; i++ {
			fl.filtered[i] = ev.GetPermutation(i)
		}

		perm := sortWithPermutations(fl.filtered, from, to, func(a, b int) int {
			return a - b
		})

		b.NextPermutation(from, to, perm)
	}
}

func (fl *FilteredList[T]) onAddRemove(ev ListChangeEvent[T]) {
	b := fl.getBuilder()
	defer b.End()

	fl.ensureSize(fl.src.Len())

	added := ev.AddedSlice()
	removed := ev.RemovedSlice()
	from := fl.findPosition(ev.From())
	to := fl.findPosition(ev.From() + len(removed))

	for i := from; i < to; i++ {
		b.NextRemove(from, removed[fl.filtered[i]-ev.From()])
	}

	fl.updateIndexes(to, len(added)-len(removed))

	fpos := from
	pos := ev.From()
	spos := 0

	for spos = pos; fpos < to && spos < ev.To(); spos++ {
		value := fl.src.Get(spos)

		if fl.predicate(value) {
			fl.filtered[fpos] = spos

			b.NextAdd(fpos, fpos+1)

			fpos++
		}
	}

	if fpos < to {
		copy(fl.filtered[fpos:], fl.filtered[to:fl.size-to])
		fl.size -= to - fpos
	} else {
		for it := spos; spos < ev.To(); it++ {
			value := fl.src.Get(it)

			if fl.predicate(value) {
				copy(fl.filtered[fpos+1:], fl.filtered[fpos:fl.size-fpos])

				fl.filtered[fpos] = it

				b.NextAdd(fpos, fpos+1)

				fpos++
				fl.size++
			}
		}
	}
}

func (fl *FilteredList[T]) onUpdated(ev ListChangeEvent[T]) {

	panic("not implemented yet")
}
