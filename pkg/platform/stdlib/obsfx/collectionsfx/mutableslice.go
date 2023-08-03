package collectionsfx

import (
	"reflect"

	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type MutableSlice[T any] struct {
	ObservableListBase[T]

	builder *ListChangeBuilder[T]
	binding obsfx.Binding[ObservableList[T]]

	slice []T
}

func (o *MutableSlice[T]) getBuilder() *ListChangeBuilder[T] {
	if o.builder == nil {
		o.builder = &ListChangeBuilder[T]{
			List:               o,
			ObservableListBase: &o.ObservableListBase,
		}
	}

	return o.builder
}

func (o *MutableSlice[T]) Clear() {
	if o.binding != nil {
		o.Unbind()
	}
	o.RemoveCount(0, o.Len())
}

func (o *MutableSlice[T]) ReplaceAll(items ...T) {
	b := o.begin()
	defer b.End()

	o.Clear()
	o.AddAll(items...)
}

func (o *MutableSlice[T]) AddAll(items ...T) int {
	return o.InsertAll(o.Len(), items...)
}

func (o *MutableSlice[T]) begin() *ListChangeBuilder[T] {
	b := o.getBuilder()
	b.Begin()
	return b
}

func (o *MutableSlice[T]) InsertAll(index int, items ...T) int {
	b := o.begin()
	defer b.End()

	count := len(items)
	to := index + count

	o.slice = slices.Insert(o.slice, index, items...)

	b.NextAdd(index, to)

	return index
}

func (o *MutableSlice[T]) Add(value T) int {
	return o.InsertAt(o.Len(), value)
}

func (o *MutableSlice[T]) InsertAt(index int, value T) int {
	return o.InsertAll(index, value)
}

func (o *MutableSlice[T]) Get(index int) T {
	return o.slice[index]
}

func (o *MutableSlice[T]) RawGet(index int) any {
	return o.Get(index)
}

func (o *MutableSlice[T]) Set(index int, value T) {
	old := o.slice[index]

	if reflect.DeepEqual(old, value) {
		return
	}

	b := o.begin()
	defer b.End()

	o.slice[index] = value

	o.builder.NextSet(index, old)
}

func (o *MutableSlice[T]) Swap(i, j int) {
	b := o.begin()
	defer b.End()

	tmp := o.Get(i)
	o.Set(i, o.Get(j))
	o.Set(j, tmp)
}

func (o *MutableSlice[T]) Remove(value T) bool {
	index := slices.IndexFunc[T](o.slice, func(t T) bool {
		return reflect.DeepEqual(t, value)
	})

	if index == -1 {
		return false
	}

	o.RemoveAt(index)

	return true
}

func (o *MutableSlice[T]) RemoveAt(index int) {
	o.RemoveCount(index, 1)
}

func (o *MutableSlice[T]) RemoveCount(index int, count int) {
	b := o.begin()
	defer b.End()

	from := index
	to := index + count

	removed := slices.Clone(o.slice[from:to])
	o.slice = slices.Delete(o.slice, from, to)

	b.NextRemoveRange(from, removed)
}

func (o *MutableSlice[T]) Slice() []T {
	return o.slice
}

func (o *MutableSlice[T]) SubSlice(from, to int) []T {
	return o.slice[from:to]
}

func (o *MutableSlice[T]) Len() int {
	return len(o.slice)
}

func (o *MutableSlice[T]) Contains(value T) bool {
	return o.IndexOf(value) != -1
}

func (o *MutableSlice[T]) IndexOf(value T) int {
	index := slices.IndexFunc[T](o.slice, func(t T) bool {
		return reflect.DeepEqual(t, value)
	})

	return index
}

func (o *MutableSlice[T]) Iterator() Iterator[T] {
	return o.ListIterator()
}

func (o *MutableSlice[T]) ListIterator() ListIterator[T] {
	return &listIterator[T]{l: o}
}

func (o *MutableSlice[T]) Bind(binding obsfx.Binding[ObservableList[T]]) {
	if o.binding != nil {
		o.Unbind()
	}

	o.binding = binding
}

func (o *MutableSlice[T]) Unbind() {
	if o.binding != nil {
		o.binding.Close()
		o.binding = nil
	}
}

type listIterator[T any] struct {
	l     ObservableList[T]
	index int
}

func (l *listIterator[T]) Index() int {
	return l.index - 1
}

func (l *listIterator[T]) Value() T {
	return l.Item()
}

func (l *listIterator[T]) Item() T {
	if l.index == 0 || l.index > l.l.Len() {
		panic("out of range")
	}

	return l.l.Get(l.index - 1)
}

func (l *listIterator[T]) Next() bool {
	if l.index >= l.l.Len() {
		return false
	}

	l.index++

	return true
}

func (l *listIterator[T]) Reset() {
	l.index = 0
}
