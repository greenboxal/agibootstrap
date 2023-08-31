package vm

import "rogchap.com/v8go"

type VM struct {
	iso *v8go.Isolate
}

func NewVM() *VM {
	iso := v8go.NewIsolate()

	return &VM{
		iso: iso,
	}
}

func (vm *VM) NewIsolate() *Isolate {
	return NewIsolate(vm)
}

func (vm *VM) Close() error {
	vm.iso.Dispose()

	return nil
}
