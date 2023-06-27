package collectionsfx

import (
	"golang.org/x/exp/slices"

	obsfx2 "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

func NewSortedList[Src ObservableList[T], T any](src Src, comparator Comparator[T]) *SortedList[T] {
	size := src.Len()
	sorted := make([]sortedListElement[T], size*3/2+1)
	perm := make([]int, len(sorted))

	fl := &SortedList[T]{
		src:    src,
		sorted: sorted,
		perm:   perm,
		size:   size,
	}

	for i := 0; i < size; i++ {
		sorted[i].index = i
		sorted[i].val = src.Get(i)
		perm[i] = i
	}

	src.AddListListener(fl)

	obsfx2.ObserveChange(&fl.comparator, func(old Comparator[T], current Comparator[T]) {
		fl.doSortWithPermutationChange()
	})

	if comparator != nil {
		fl.SetComparator(comparator)
	}

	return fl
}

type Comparator[T any] func(a, b T) int

type sortedListElement[T any] struct {
	val   T
	index int
}

type SortedList[T any] struct {
	ObservableListBase[T]

	comparator obsfx2.SimpleProperty[Comparator[T]]

	src     ObservableList[T]
	builder *ListChangeBuilder[T]

	sorted []sortedListElement[T]
	perm   []int
	size   int
}

func (sl *SortedList[T]) Iterator() Iterator[T] {
	return &listIterator[T]{l: sl}
}

func (sl *SortedList[T]) SetComparator(comparator Comparator[T]) {
	sl.comparator.SetValue(comparator)
}

func (sl *SortedList[T]) Source() ObservableList[T] {
	return sl.src
}

func (sl *SortedList[T]) GetSourceIndex(index int) int {
	return sl.sorted[index].index
}

func (sl *SortedList[T]) GetViewIndex(index int) int {
	return sl.perm[index]
}

func (sl *SortedList[T]) RawGet(index int) any {
	return sl.Get(index)
}

func (sl *SortedList[T]) Get(index int) T {
	if index >= sl.size {
		panic("out of bounds")
	}

	return sl.sorted[index].val
}

func (sl *SortedList[T]) Slice() []T {
	return sl.SubSlice(0, sl.Len())
}

func (sl *SortedList[T]) SubSlice(from, to int) []T {
	result := make([]T, to-from)

	for i := range result {
		result[i] = sl.Get(from + i)
	}

	return result
}

func (sl *SortedList[T]) Len() int {
	return sl.size
}

func (sl *SortedList[T]) Contains(value T) bool {
	return sl.IndexOf(value) >= 0
}

func (sl *SortedList[T]) IndexOf(value T) int {
	index := sl.src.IndexOf(value)

	if index == -1 {
		return -1
	}

	return sl.GetViewIndex(index)
}

func (sl *SortedList[T]) begin() *ListChangeBuilder[T] {
	if sl.builder == nil {
		sl.builder = &ListChangeBuilder[T]{
			List:               sl,
			ObservableListBase: &sl.ObservableListBase,
		}
	}

	sl.builder.Begin()

	return sl.builder
}

func (sl *SortedList[T]) OnListChanged(ev ListChangeEvent[T]) {
	b := sl.begin()
	defer b.End()

	for ev.Next() {
		if ev.WasPermutated() {
			sl.onPermutation(ev)
		} else if ev.WasUpdated() {
			sl.onUpdate(ev)
		} else {
			sl.onAddRemove(ev)
		}
	}
}

func (sl *SortedList[T]) Close() {
	sl.src.RemoveListListener(sl)
}

func (sl *SortedList[T]) onPermutation(change ListChangeEvent[T]) {
	for i := 0; i < sl.size; i++ {
		p := change.GetPermutation(sl.sorted[i].index)

		sl.sorted[i].index = p
		sl.perm[p] = i
	}
}

func (sl *SortedList[T]) onUpdate(change ListChangeEvent[T]) {
	b := sl.begin()
	defer b.End()

	perm := sortWithPermutations(sl.sorted, 0, sl.size, sl.elementComparator)

	for i := 0; i < sl.size; i++ {
		sl.perm[sl.sorted[i].index] = i
	}

	b.NextPermutation(0, sl.size, perm)

	for i := change.From(); i < change.To(); i++ {
		b.NextUpdate(sl.perm[i])
	}
}

func (sl *SortedList[T]) onAddRemove(change ListChangeEvent[T]) {
	removed := change.RemovedSlice()

	if change.From() == 0 && len(removed) == sl.Len() {
		sl.removeAllFromMapping()
	} else {
		for _, v := range removed {
			sl.removeFromMapping(change.From(), v)
		}
	}

	if sl.size == 0 {
		sl.setAllToMapping(change.List(), change.To())
	} else {
		for i, v := range change.AddedSlice() {
			sl.insertToMapping(change.From()+i, v)
		}
	}
}

func (sl *SortedList[T]) removeAllFromMapping() {
	b := sl.begin()
	defer b.End()

	removed := make([]T, sl.size)

	for i := 0; i < sl.size; i++ {
		removed[i] = sl.Get(i)
	}

	sl.size = 0

	b.NextRemoveRange(0, removed)
}

func (sl *SortedList[T]) removeFromMapping(index int, removed T) {
	b := sl.begin()
	defer b.End()

	pos := sl.perm[index]

	sl.sorted = slices.Delete(sl.sorted, pos, pos+1)
	sl.perm = slices.Delete(sl.perm, pos, pos+1)

	sl.size--
	sl.sorted[sl.size].index = -1

	sl.updateIndices(index+1, pos, -1)
}

func (sl *SortedList[T]) setAllToMapping(list ObservableList[T], to int) {
	b := sl.begin()
	defer b.End()

	sl.ensureSize(to)
	sl.size = to

	for i := 0; i < to; i++ {
		sl.sorted[i] = sortedListElement[T]{
			index: i,
			val:   list.Get(i),
		}
	}

	perm := sortWithPermutations(sl.sorted, 0, sl.size, sl.elementComparator)
	copy(sl.perm, perm[:sl.size])

	b.NextAdd(0, sl.size)
}

func (sl *SortedList[T]) insertToMapping(index int, value T) {
	b := sl.begin()
	defer b.End()

	pos := sl.findPosition(index)

	if pos < 0 {
		pos = ^pos
	}

	sl.ensureSize(sl.size + 1)
	sl.updateIndices(index, pos, 1)

	sl.sorted = slices.Insert(sl.sorted, pos, sortedListElement[T]{
		val:   value,
		index: index,
	})

	sl.perm = slices.Insert(sl.perm, index, pos)

	sl.size++

	b.NextAdd(pos, pos+1)
}

func (sl *SortedList[T]) updateIndices(from int, viewFrom int, delta int) {
	for i := 0; i < sl.size; i++ {
		if sl.sorted[i].index >= from {
			sl.sorted[i].index += delta
		}

		if sl.perm[i] >= viewFrom {
			sl.perm[i] += delta
		}
	}
}

func (sl *SortedList[T]) findPosition(index int) int {
	if len(sl.sorted) == 0 {
		return 0
	}

	if index == 0 {
		return 0
	}

	index, ok := slices.BinarySearchFunc(sl.sorted[:sl.size], index, func(s sortedListElement[T], i int) int {
		return s.index - i
	})

	if !ok {
		return ^index
	}

	return index
}

func (sl *SortedList[T]) ensureSize(size int) {
	if len(sl.sorted) < size {
		filtered := make([]sortedListElement[T], size*3/2+1)
		copy(filtered, sl.sorted[:sl.size])
		sl.sorted = filtered

		perm := make([]int, len(filtered))
		copy(perm, sl.perm[:sl.size])
		sl.perm = perm
	}
}

func (sl *SortedList[T]) doSortWithPermutationChange() {
	b := sl.begin()
	defer b.End()

	perm := sortWithPermutations(sl.sorted, 0, sl.size, sl.elementComparator)

	for i := 0; i < sl.size; i++ {
		sl.perm[sl.sorted[i].index] = i
	}

	b.NextPermutation(0, sl.size, perm)
}

func (sl *SortedList[T]) elementComparator(a sortedListElement[T], b sortedListElement[T]) int {
	return sl.comparator.Value()(a.val, b.val)
}
