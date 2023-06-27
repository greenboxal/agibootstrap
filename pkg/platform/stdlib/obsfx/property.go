package obsfx

import (
	"reflect"
)

type IProperty interface {
	Property() IProperty
	Close()
}

type ReadOnlyProperty[T any] interface {
	IProperty
	ObservableValue[T]
}

type Property[T any] interface {
	IProperty
	ReadOnlyProperty[T]

	IsBound() bool

	Bind(observable ObservableValue[T])
	Unbind()

	BindBidirectional(other Property[T])
	UnbindBidirectional()

	SetValue(value T)
}

type SimpleProperty[T any] struct {
	listener *simplePropertyBindingListener[T]
	helper   ExpressionHelper[T]

	value         T
	valid         bool
	bindingTarget ObservableValue[T]
}

func (o *SimpleProperty[T]) IsBound() bool {
	return o.bindingTarget != nil
}

func (o *SimpleProperty[T]) Property() IProperty {
	return o
}

func (o *SimpleProperty[T]) RawValue() any {
	return o.Value()
}

func (o *SimpleProperty[T]) Value() T {
	o.valid = true

	if o.bindingTarget != nil {
		return o.bindingTarget.Value()
	}

	return o.value
}

func (o *SimpleProperty[T]) SetValue(value T) {
	if o.IsBound() {
		panic("cannot SetValue on bound target")
	}

	o.doSetValue(value)
}

func (o *SimpleProperty[T]) Bind(observable ObservableValue[T]) {
	if observable == nil {
		panic("cannot bind to nil observable")
	}

	if o.bindingTarget == observable {
		return
	}

	if o.bindingTarget != nil {
		o.lockedUnbind()
	}

	if o.listener == nil {
		o.listener = &simplePropertyBindingListener[T]{prop: o}
	}

	o.bindingTarget = observable

	observable.AddListener(o.listener)

	o.value = observable.Value()

	o.markInvalid()
}

func (o *SimpleProperty[T]) Unbind() {
	o.unlockedUnbind()
}

func (o *SimpleProperty[T]) BindBidirectional(other Property[T]) {
	b := &bidirectionalBinding[T]{
		prop1: other,
		prop2: o,
	}

	o.AddListener(b)
	other.AddListener(b)

	b.OnInvalidated(other)
}

func (o *SimpleProperty[T]) UnbindBidirectional() {
	o.unlockedUnbind()
}

func (o *SimpleProperty[T]) doSetValue(value T) {
	if reflect.DeepEqual(o.value, value) {
		return
	}

	o.value = value
	o.valid = true

	o.markInvalid()
}

func (o *SimpleProperty[T]) lockedUnbind() {
	if o.bindingTarget != nil {
		o.value = o.bindingTarget.Value()
		o.bindingTarget.RemoveListener(o.listener)
		o.bindingTarget = nil
	}
}

func (o *SimpleProperty[T]) unlockedUnbind() {
	o.lockedUnbind()
}

func (o *SimpleProperty[T]) Close() {
	o.lockedUnbind()
}

func (o *SimpleProperty[T]) markInvalid() {
	if o.valid {
		o.valid = false

		if o.helper != nil {
			o.helper.OnInvalidated(o)
		}
	}
}

func (o *SimpleProperty[T]) AddListener(listener InvalidationListener) {
	if o.helper == nil {
		o.helper = NewSingleInvalidationExpressionHelper[T](listener)
	} else {
		o.helper = o.helper.AddListener(listener)
	}
}

func (o *SimpleProperty[T]) RemoveListener(listener InvalidationListener) {
	if o.helper != nil {
		o.helper.RemoveListener(listener)
	}
}

func (o *SimpleProperty[T]) AddChangeListener(listener ChangeListener[T]) {
	if o.helper == nil {
		o.helper = NewSingleChangeListenerExpressionHelper[T](listener)
	} else {
		o.helper = o.helper.AddChangeListener(listener)
	}
}

func (o *SimpleProperty[T]) RemoveChangeListener(listener ChangeListener[T]) {
	if o.helper != nil {
		o.helper.RemoveChangeListener(listener)
	}
}

type simplePropertyBindingListener[T any] struct{ prop *SimpleProperty[T] }

func (s simplePropertyBindingListener[T]) OnInvalidated(prop Observable) {
	s.prop.markInvalid()
}

type StringProperty = SimpleProperty[string]
type BoolProperty = SimpleProperty[bool]
type FloatProperty = SimpleProperty[float32]
type DoubleProperty = SimpleProperty[float64]
type IntProperty = SimpleProperty[int]
