package collectionsfx

import (
	"reflect"

	obsfx2 "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type BasicListListener interface {
	OnListChangedRaw(ev BasicListChangeEvent)
}

type ListListener[T any] interface {
	OnListChanged(ev ListChangeEvent[T])
}

type OnListChangedFunc[T any] func(ev ListChangeEvent[T])

func (o OnListChangedFunc[T]) OnListChanged(ev ListChangeEvent[T]) {
	o(ev)
}

type OnListChangedRawFunc[T any] func(ev BasicListChangeEvent)

func (o OnListChangedRawFunc[T]) OnListChangedRaw(ev BasicListChangeEvent) {
	o(ev)
}

func (o OnListChangedRawFunc[T]) OnListChanged(ev ListChangeEvent[T]) {
	o(ev)
}

type BasicObservableList interface {
	obsfx2.Observable

	RawGet(i int) interface{}
	Len() int
	RuntimeElementType() reflect.Type

	AddBasicListListener(listener BasicListListener)
	RemoveBasicListListener(listener BasicListListener)
}

type ObservableList[T any] interface {
	BasicObservableList

	Iterable[T]

	Get(index int) T
	Slice() []T
	SubSlice(from, to int) []T
	Contains(value T) bool
	IndexOf(value T) int

	AddListListener(listener ListListener[T])
	RemoveListListener(listener ListListener[T])
}

type ModifiableObservableList[T any] interface {
	ObservableList[T]

	Add(value T) int
	AddAll(value ...T) int
	InsertAt(index int, value T) int
	InsertAll(index int, items ...T) int
	ReplaceAll(items ...T)
	Set(index int, value T)
	Swap(i, j int)
	Remove(value T) bool
	RemoveAt(index int)
	RemoveCount(index int, count int)
	Clear()

	Bind(binding obsfx2.Binding[ObservableList[T]])
	Unbind()
}

type ObservableListBase[T any] struct {
	helper listExpressionHelper[T]
}

func (o *ObservableListBase[T]) RuntimeElementType() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func (o *ObservableListBase[T]) AddBasicListListener(listener BasicListListener) {
	o.AddListListener(OnListChangedRawFunc[T](listener.OnListChangedRaw))
}

func (o *ObservableListBase[T]) RemoveBasicListListener(listener BasicListListener) {
	o.RemoveListListener(OnListChangedRawFunc[T](listener.OnListChangedRaw))
}

func (o *ObservableListBase[T]) AddListListener(listener ListListener[T]) {
	if o.helper == nil {
		o.helper = &singleListListenerExpressionHelper[T]{listener: listener}
	} else {
		o.helper = o.helper.AddListListener(listener)
	}
}

func (o *ObservableListBase[T]) RemoveListListener(listener ListListener[T]) {
	if o.helper != nil {
		o.helper = o.helper.RemoveListListener(listener)
	}
}

func (o *ObservableListBase[T]) AddListener(listener obsfx2.InvalidationListener) {
	if o.helper == nil {
		o.helper = &singleInvalidationListExpressionHelper[T]{listener: listener}
	} else {
		o.helper = o.helper.AddListener(listener)
	}
}

func (o *ObservableListBase[T]) RemoveListener(listener obsfx2.InvalidationListener) {
	if o.helper != nil {
		o.helper = o.helper.RemoveListener(listener)
	}
}

func (o *ObservableListBase[T]) FireListChanged(ev ListChangeEvent[T]) {
	if o.helper != nil {
		o.helper.OnListChanged(ev)
	}
}
