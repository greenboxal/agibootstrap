package obsfx

type ObjectBinding[T, V any] struct {
	ObservableValueBase[V]

	src    ObservableValue[T]
	mapper func(T) V

	value V
	valid bool
}

func NewObjectBinding[T, V any](src ObservableValue[T], mapper func(T) V) *ObjectBinding[T, V] {
	ob := &ObjectBinding[T, V]{
		src:    src,
		mapper: mapper,
	}

	src.AddListener(ob)

	return ob
}

func (b *ObjectBinding[T, V]) Dependencies() []Observable {
	return []Observable{b.src}
}

func (b *ObjectBinding[T, V]) Invalidate() {
	b.markInvalid()
}

func (b *ObjectBinding[T, V]) IsValid() bool {
	return b.valid
}

func (o *ObjectBinding[T, V]) RawValue() any {
	return o.Value()
}

func (b *ObjectBinding[T, V]) Value() V {
	if !b.valid {
		b.value = b.mapper(b.src.Value())
		b.valid = true
	}

	return b.value
}

func (b *ObjectBinding[T, V]) OnInvalidated(prop Observable) {
	b.markInvalid()
}

func (b *ObjectBinding[T, V]) Close() {
	for _, dep := range b.Dependencies() {
		dep.RemoveListener(b)
	}
}

func (b *ObjectBinding[T, V]) markInvalid() {
	if b.valid {
		b.valid = false

		b.ObservableValueBase.OnInvalidated(b)

		if !b.valid {
			b.value = EmptyValue[V]()
		}
	}
}
