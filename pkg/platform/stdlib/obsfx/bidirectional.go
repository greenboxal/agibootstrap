package obsfx

type bidirectionalBinding[T any] struct {
	prop1    Property[T]
	prop2    Property[T]
	updating bool
}

func (b *bidirectionalBinding[T]) OnInvalidated(observable Observable) {
	if b.updating {
		return
	}

	if b.prop1 == nil || b.prop2 == nil {
		if b.prop1 != nil {
			b.prop1.RemoveListener(b)
		}

		if b.prop2 != nil {
			b.prop2.RemoveListener(b)
		}

		return
	}

	defer func() {
		b.updating = false
	}()

	b.updating = true

	if observable == b.prop1 {
		b.prop2.SetValue(b.prop1.Value())
	} else if observable == b.prop2 {
		b.prop1.SetValue(b.prop1.Value())
	}
}

func (b *bidirectionalBinding[T]) Close() {
	if b.prop1 != nil {
		b.prop1.RemoveListener(b)
		b.prop1 = nil
	}

	if b.prop2 != nil {
		b.prop2.RemoveListener(b)
		b.prop2 = nil
	}
}
