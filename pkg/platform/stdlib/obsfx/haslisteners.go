package obsfx

import (
	"reflect"
	"sync"
)

type HasListenersBase[T any] struct {
	m         sync.RWMutex
	empty     T
	listeners []T
}

func (o *HasListenersBase[T]) AddListener(listener T) {
	o.m.Lock()
	defer o.m.Unlock()

	for i, v := range o.listeners {
		if reflect.DeepEqual(v, o.empty) {
			o.listeners[i] = listener
			return
		}
	}

	o.listeners = append(o.listeners, listener)
}

func (o *HasListenersBase[T]) RemoveListener(listener T) {
	o.m.Lock()
	defer o.m.Unlock()

	for i, v := range o.listeners {
		if reflect.DeepEqual(listener, v) {
			o.listeners[i] = o.empty
		}
	}
}

func (o *HasListenersBase[T]) ForEachListener(do func(l T) bool) {
	o.m.RLock()
	defer o.m.RUnlock()

	for _, l := range o.listeners {
		do(l)
	}
}

func (o *HasListenersBase[T]) Close() {
	o.listeners = nil
}
