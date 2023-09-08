package vm

import (
	"reflect"
	"sync"
	"sync/atomic"

	"rogchap.com/v8go"
)

type ObjectHandleTable struct {
	mu sync.RWMutex

	wrappers map[uint64]*v8go.Value
	handles  map[uint64]reflect.Value
	objects  map[any]uint64

	nextHandle atomic.Uint64
}

func NewObjectHandleTable() *ObjectHandleTable {
	return &ObjectHandleTable{
		wrappers: map[uint64]*v8go.Value{},
		handles:  map[uint64]reflect.Value{},
		objects:  map[any]uint64{},
	}
}

func (t *ObjectHandleTable) Add(value reflect.Value, factory func(handle int) (*v8go.Value, error)) (uint64, *v8go.Value, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	ptr := value.Interface()

	if h, ok := t.objects[ptr]; ok {
		return h, t.wrappers[h], nil
	}
	handle := t.nextHandle.Add(1)
	wrapper, err := factory(int(handle))

	if err != nil {
		return 0, nil, err
	}

	t.handles[handle] = value
	t.wrappers[handle] = wrapper
	t.objects[ptr] = handle

	return handle, wrapper, nil
}

func (t *ObjectHandleTable) Remove(handle uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	value, ok := t.handles[handle]

	if !ok {
		return
	}

	ptr := value.UnsafePointer()

	delete(t.handles, handle)
	delete(t.wrappers, handle)
	delete(t.objects, ptr)
}

func (t *ObjectHandleTable) Get(handle uint64) reflect.Value {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.handles[handle]
}

func (t *ObjectHandleTable) LookupHandle(value reflect.Value) (uint64, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	h, ok := t.objects[value.UnsafePointer()]

	return h, ok
}
