package obsfx

type ChangeListener[T any] interface {
	OnChanged(observable ObservableValue[T], old, current T)
}

type OnChangedFunc[T any] func(prop ObservableValue[T], old, current T)

func (o OnChangedFunc[T]) OnChanged(prop ObservableValue[T], old, current T) {
	o(prop, old, current)
}

type InvalidationListener interface {
	OnInvalidated(observable Observable)
}

type OnInvalidatedFunc func(src Observable)

func (o OnInvalidatedFunc) OnInvalidated(src Observable) {
	o(src)
}

type Observable interface {
	AddListener(listener InvalidationListener)
	RemoveListener(listener InvalidationListener)
}

type RawObservableValue interface {
	Observable

	RawValue() any
}

type ObservableValue[T any] interface {
	RawObservableValue

	Value() T

	AddChangeListener(listener ChangeListener[T])
	RemoveChangeListener(listener ChangeListener[T])
}

type ObservableValueBase[T any] struct {
	helper ExpressionHelper[T]
}

func (o *ObservableValueBase[T]) OnInvalidated(obs Observable) {
	if o.helper != nil {
		o.helper.OnInvalidated(obs)
	}
}

func (o *ObservableValueBase[T]) AddListener(listener InvalidationListener) {
	if o.helper == nil {
		o.helper = NewSingleInvalidationExpressionHelper[T](listener)
	} else {
		o.helper = o.helper.AddListener(listener)
	}
}

func (o *ObservableValueBase[T]) RemoveListener(listener InvalidationListener) {
	if o.helper != nil {
		o.helper.RemoveListener(listener)
	}
}

func (o *ObservableValueBase[T]) AddChangeListener(listener ChangeListener[T]) {
	if o.helper == nil {
		o.helper = NewSingleChangeListenerExpressionHelper[T](listener)
	} else {
		o.helper = o.helper.AddChangeListener(listener)
	}
}

func (o *ObservableValueBase[T]) RemoveChangeListener(listener ChangeListener[T]) {
	if o.helper != nil {
		o.helper.RemoveChangeListener(listener)
	}
}
