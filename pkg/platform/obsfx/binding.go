package obsfx

import (
	"reflect"
)

type Binding[V any] interface {
	ObservableValue[V]

	Dependencies() []Observable
	Invalidate()
	IsValid() bool

	Close()
}

func ObserveInvalidation[Src Observable](src Src, cb func()) {
	src.AddListener(OnInvalidatedFunc(func(prop Observable) {
		cb()
	}))
}

func ObserveValue[Src ObservableValue[T], T any](src Src, cb func(T)) {
	src.AddChangeListener(OnChangedFunc[T](func(prop ObservableValue[T], old, current T) {
		cb(current)
	}))
}

func ObserveChange[Src ObservableValue[T], T any](src Src, cb func(T, T)) {
	src.AddChangeListener(OnChangedFunc[T](func(prop ObservableValue[T], old, current T) {
		cb(old, current)
	}))
}

func BindFunc[Src ObservableValue[T], T any](f func(T), src Src) Binding[T] {
	b := NewObjectBinding[T, T](src, func(t T) T {
		return t
	})

	ObserveValue(b, f)

	b.Invalidate()

	f(b.Value())

	return b
}

func Bind[Src ObservableValue[T], T, V any](src Src, mapper func(T) V) Binding[V] {
	b := NewObjectBinding[T, V](src, mapper)

	b.Invalidate()

	return b
}

func BindAny[V any](src RawObservableValue, mapper func(any) V) Binding[V] {
	srcAny := src.(ObservableValue[any])

	b := NewObjectBinding[any, V](srcAny, mapper)

	b.Invalidate()

	return b
}

func Select[Src ObservableValue[V], V, Result ObservableValue[T], T any](src Src, mapper func(V) Result) Binding[T] {
	var step SelectStep

	existing, isSelect := reflect.ValueOf(src).Interface().(*SelectBinding[V])

	step = func(a any) Observable {
		v := a.(V)

		return mapper(v)
	}

	if isSelect {
		var steps []SelectStep

		steps = append(steps, existing.steps...)
		steps = append(steps, step)

		return NewSelectBinding[T](existing.props[0], steps)
	}

	return NewSelectBinding[T](src, []SelectStep{step})
}

func EmptyValue[V any]() V {
	var v V
	return v
}
