package obsfx

func Just[V any](v V) ObservableValue[V] {
	return immutableObservableValue[V]{v: v}
}

type Invalidatable interface {
	Invalidate()
}

type ExpressionBinding[V any] struct {
	ObservableValueBase[V]

	fn    func() V
	value V
	valid bool
}

func (e *ExpressionBinding[V]) RawValue() any {
	return e.Value()
}

func (e *ExpressionBinding[V]) Value() V {
	if e.valid == false {
		e.value = e.fn()
		e.valid = true
	}

	return e.value
}

func (e *ExpressionBinding[V]) Invalidate() {
	if e.valid {
		e.valid = false

		e.OnInvalidated(e)
	}
}

func BindExpression[V any](fn func() V, deps ...Observable) *ExpressionBinding[V] {
	expr := &ExpressionBinding[V]{fn: fn}

	for _, dep := range deps {
		ObserveInvalidation(dep, func() {
			expr.Invalidate()
		})
	}

	return expr
}

func DropRet[T, R any](f func(T) R) func(T) {
	return func(t T) {
		f(t)
	}
}

type immutableObservableValue[V any] struct {
	v V
}

func (i immutableObservableValue[V]) RawValue() any {
	return i.v
}

func (i immutableObservableValue[V]) Value() V {
	return i.v
}

func (i immutableObservableValue[V]) AddListener(listener InvalidationListener) {
}

func (i immutableObservableValue[V]) RemoveListener(listener InvalidationListener) {
}

func (i immutableObservableValue[V]) AddChangeListener(listener ChangeListener[V]) {
}

func (i immutableObservableValue[V]) RemoveChangeListener(listener ChangeListener[V]) {
}
