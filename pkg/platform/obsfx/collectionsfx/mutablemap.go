package collectionsfx

import (
	"reflect"
	"sync"

	"github.com/greenboxal/agibootstrap/pkg/platform/obsfx"
)

type MutableMap[K comparable, V any] struct {
	m      sync.RWMutex
	kv     map[K]V
	helper mapExpressionHelper[K, V]
}

func (o *MutableMap[K, V]) Iterator() Iterator[KeyValuePair[K, V]] {
	return o.MapIterator()
}

func (o *MutableMap[K, V]) MapIterator() MapIterator[K, V] {
	return &mutableMapIterator[K, V]{m: o}
}

func (o *MutableMap[K, V]) Clear() {
	for k := range o.kv {
		o.Remove(k)
	}
}

func (o *MutableMap[K, V]) Has(key K) bool {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.kv == nil {
		o.kv = map[K]V{}
	}

	_, ok := o.kv[key]

	return ok
}

func (o *MutableMap[K, V]) Get(key K) (V, bool) {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.kv == nil {
		o.kv = map[K]V{}
	}

	v, ok := o.kv[key]

	return v, ok
}

func (o *MutableMap[K, V]) Set(key K, value V) {
	doSet := func() (old V, added bool, removed bool) {
		o.m.Lock()
		defer o.m.Unlock()

		if o.kv == nil {
			o.kv = map[K]V{}
		}

		old, ok := o.kv[key]

		if reflect.DeepEqual(old, value) {
			return old, false, false
		}

		o.kv[key] = value

		return old, true, ok
	}

	if old, added, removed := doSet(); added || removed {
		o.fireListeners(MapChangeEvent[K, V]{
			Map:          o,
			Key:          key,
			ValueAdded:   value,
			ValueRemoved: old,
			WasAdded:     added,
			WasRemoved:   removed,
		})
	}
}

func (o *MutableMap[K, V]) Remove(key K) bool {
	doRemove := func() (empty V, ok bool) {
		o.m.Lock()
		defer o.m.Unlock()

		if o.kv == nil {
			o.kv = map[K]V{}
		}

		existing, ok := o.kv[key]

		if !ok {
			return empty, false
		}

		delete(o.kv, key)

		return existing, true
	}

	if value, ok := doRemove(); ok {
		o.fireListeners(MapChangeEvent[K, V]{
			Map:          o,
			Key:          key,
			ValueRemoved: value,
			WasRemoved:   true,
		})

		return true
	}

	return false
}

func (o *MutableMap[K, V]) ensureMap() map[K]V {
	o.m.Lock()
	defer o.m.Unlock()

	if o.kv == nil {
		o.kv = map[K]V{}
	}

	return o.kv
}

func (o *MutableMap[K, V]) Map() map[K]V {
	return o.ensureMap()
}

func (o *MutableMap[K, V]) Len() int {
	return len(o.kv)
}

func (o *MutableMap[K, V]) Keys() []K {
	o.m.RLock()
	defer o.m.RUnlock()

	keys := make([]K, 0, len(o.kv))

	for k := range o.kv {
		keys = append(keys, k)
	}

	return keys
}

func (o *MutableMap[K, V]) Values() []V {
	o.m.RLock()
	defer o.m.RUnlock()

	values := make([]V, 0, len(o.kv))

	for _, v := range o.kv {
		values = append(values, v)
	}

	return values
}

func (o *MutableMap[K, V]) fireListeners(ev MapChangeEvent[K, V]) {
	if o.helper != nil {
		o.helper.OnMapChanged(ev)
	}
}

func (o *MutableMap[K, V]) AddListener(listener obsfx.InvalidationListener) {
	o.m.Lock()
	defer o.m.Unlock()

	if o.helper == nil {
		o.helper = &singleInvalidationMapExpressionHelper[K, V]{listener: listener}
	} else {
		o.helper = o.helper.AddListener(listener)
	}
}

func (o *MutableMap[K, V]) RemoveListener(listener obsfx.InvalidationListener) {
	o.m.Lock()
	defer o.m.Unlock()

	if o.helper != nil {
		o.helper = o.helper.RemoveListener(listener)
	}
}

func (o *MutableMap[K, V]) AddMapListener(listener MapListener[K, V]) {
	o.m.Lock()
	defer o.m.Unlock()

	if o.helper == nil {
		o.helper = &singleMapListenerExpressionHelper[K, V]{listener: listener}
	} else {
		o.helper = o.helper.AddMapListener(listener)
	}
}

func (o *MutableMap[K, V]) RemoveMapListener(listener MapListener[K, V]) {
	o.m.Lock()
	defer o.m.Unlock()

	if o.helper != nil {
		o.helper = o.helper.RemoveMapListener(listener)
	}
}

type mutableMapIterator[K comparable, V any] struct {
	m     *MutableMap[K, V]
	index int
	keys  []K
}

func (it *mutableMapIterator[K, V]) Item() KeyValuePair[K, V] {
	if it.index > len(it.keys) {
		panic("out of range")
	}

	k := it.keys[it.index-1]
	v, _ := it.m.Get(k)

	return KeyValuePair[K, V]{
		Key:   k,
		Value: v,
	}
}

func (it *mutableMapIterator[K, V]) Next() bool {
	if it.index == 0 && it.keys == nil {
		it.keys = it.m.Keys()
	}

	if it.index >= len(it.keys) {
		it.keys = nil
		return false
	}

	it.index++

	return true
}

func (it *mutableMapIterator[K, V]) Reset() {
	it.index = 0
	it.keys = nil
}

func (it *mutableMapIterator[K, V]) Pair() (K, V) {
	return it.Key(), it.Value()
}

func (it *mutableMapIterator[K, V]) Key() K {
	return it.Item().Key
}

func (it *mutableMapIterator[K, V]) Value() V {
	return it.Item().Value
}
