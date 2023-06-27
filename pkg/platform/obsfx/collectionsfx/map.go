package collectionsfx

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/obsfx"
)

type MapChangeEvent[K comparable, V any] struct {
	Map          ObservableMap[K, V]
	Key          K
	ValueAdded   V
	ValueRemoved V
	WasAdded     bool
	WasRemoved   bool
}

type MapListener[K comparable, V any] interface {
	OnMapChanged(ev MapChangeEvent[K, V])
}

type OnMapChangedFunc[K comparable, V any] func(ev MapChangeEvent[K, V])

func (o OnMapChangedFunc[K, V]) OnMapChanged(ev MapChangeEvent[K, V]) {
	o(ev)
}

type KeyValuePair[K comparable, V any] struct {
	Key   K
	Value V
}

type MapIterator[K comparable, V any] interface {
	Iterator[KeyValuePair[K, V]]

	Key() K
	Value() V
	Pair() (K, V)
}

type ObservableMap[K comparable, V any] interface {
	obsfx.Observable

	Iterable[KeyValuePair[K, V]]

	Has(key K) bool
	Get(key K) (V, bool)
	Map() map[K]V
	Len() int

	Keys() []K
	Values() []V

	MapIterator() MapIterator[K, V]

	AddMapListener(listener MapListener[K, V])
	RemoveMapListener(listener MapListener[K, V])
}

type ModifiableObservableMap[K comparable, V any] interface {
	ObservableMap[K, V]

	Set(key K, value V)
	Remove(key K) bool
	Clear()
}

func ObserveMap[Src ObservableMap[K, V], K comparable, V any](src Src, cb func(ev MapChangeEvent[K, V])) {
	src.AddMapListener(OnMapChangedFunc[K, V](cb))

	for k, v := range src.Map() {
		cb(MapChangeEvent[K, V]{
			Map:        src,
			Key:        k,
			ValueAdded: v,
			WasAdded:   true,
		})
	}
}
