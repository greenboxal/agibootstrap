package vm

import (
	"rogchap.com/v8go"
)

type Isolate struct {
	vm          *VM
	iso         *v8go.Isolate
	moduleCache *ModuleCache
}

func NewIsolate(vm *VM) *Isolate {
	iso := v8go.NewIsolate()

	return &Isolate{
		vm:          vm,
		iso:         iso,
		moduleCache: NewModuleCache(vm),
	}
}

func (iso *Isolate) Close() error {
	iso.iso.Dispose()

	return nil
}
