package collectionsfx

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/obsfx"
)

type SetChangeEvent[K comparable] struct {
	Set        ObservableSet[K]
	Key        K
	WasAdded   bool
	WasRemoved bool
}

type SetListener[K comparable] interface {
	OnSetChanged(ev SetChangeEvent[K])
}

type OnSetChangedFunc[K comparable] func(ev SetChangeEvent[K])

func (o OnSetChangedFunc[K]) OnSetChanged(ev SetChangeEvent[K]) {
	o(ev)
}

type ObservableSet[K comparable] interface {
	obsfx.Observable

	Iterable[K]

	Has(key K) bool
	Add(key K) bool
	Remove(key K) bool
	Slice() []K
	Len() int

	AddSetListener(listener SetListener[K])
	RemoveSetListener(listener SetListener[K])
}

func ObserveSet[Src ObservableSet[K], K comparable](src Src, cb func(ev SetChangeEvent[K])) {
	src.AddSetListener(OnSetChangedFunc[K](cb))

	for _, k := range src.Slice() {
		cb(SetChangeEvent[K]{
			Set:      src,
			Key:      k,
			WasAdded: true,
		})
	}
}

type ObservableSetBase[T comparable] struct {
	helper setExpressionHelper[T]
}

func (o *ObservableSetBase[T]) AddSetListener(listener SetListener[T]) {
	if o.helper == nil {
		o.helper = &singleSetListenerExpressionHelper[T]{listener: listener}
	} else {
		o.helper = o.helper.AddSetListener(listener)
	}
}

func (o *ObservableSetBase[T]) RemoveSetListener(listener SetListener[T]) {
	if o.helper != nil {
		o.helper = o.helper.RemoveSetListener(listener)
	}
}

func (o *ObservableSetBase[T]) AddListener(listener obsfx.InvalidationListener) {
	if o.helper == nil {
		o.helper = &singleInvalidationSetExpressionHelper[T]{listener: listener}
	} else {
		o.helper = o.helper.AddListener(listener)
	}
}

func (o *ObservableSetBase[T]) RemoveListener(listener obsfx.InvalidationListener) {
	if o.helper != nil {
		o.helper = o.helper.RemoveListener(listener)
	}
}

func (o *ObservableSetBase[T]) FireSetChanged(ev SetChangeEvent[T]) {
	if o.helper != nil {
		o.helper.OnSetChanged(ev)
	}
}
