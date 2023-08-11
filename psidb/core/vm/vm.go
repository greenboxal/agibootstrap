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

func (vm *VM) NewContext() *Context {
	ctx := v8go.NewContext(vm.iso)

	return &Context{
		ctx: ctx,
	}
}

type Context struct {
	ctx *v8go.Context
}

func (ctx *Context) Close() error {
	ctx.ctx.Close()

	return nil
}
