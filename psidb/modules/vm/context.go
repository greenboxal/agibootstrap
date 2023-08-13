package vm

import (
	"context"

	"rogchap.com/v8go"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Context struct {
	baseCtx context.Context

	iso *Isolate
	sp  inject.ServiceLocator

	ctx *v8go.Context

	modules map[string]*LiveModule
}

func NewContext(baseCtx context.Context, iso *Isolate, sp inject.ServiceLocator) *Context {
	ctx := v8go.NewContext(iso.iso)

	return &Context{
		baseCtx: baseCtx,
		iso:     iso,
		ctx:     ctx,
		sp:      sp,

		modules: map[string]*LiveModule{},
	}
}

func (vmctx *Context) Close() error {
	vmctx.ctx.Close()

	return nil
}

func (vmctx *Context) Load(ctx context.Context, m *Module) (*LiveModule, error) {
	if m.cached == nil {
		src := ModuleSource{
			Name:   m.Name,
			Source: m.Source,
		}

		cached := NewCachedModule(vmctx.iso.vm, m.Name, src)

		m.cached = cached
	}

	lm, err := NewLiveModule(vmctx, m.cached)

	if err != nil {
		return nil, err
	}

	vmctx.modules[m.Name] = lm

	return lm.Get()
}

func (vmctx *Context) Require(ctx context.Context, name string) (*LiveModule, error) {
	if m := vmctx.modules[name]; m != nil {
		return m.Get()
	}

	mod, err := vmctx.iso.moduleCache.Get(name)

	if err != nil {
		return nil, err
	}

	if mod == nil {
		g := inject.Inject[psi.Graph](vmctx.sp)
		path, err := psi.ParsePath(name)

		if err != nil {
			return nil, err
		}

		mod, err := psi.Resolve[*Module](ctx, g, path)

		if err != nil {
			return nil, err
		}

		return mod.Get(ctx)
	}

	lm, err := NewLiveModule(vmctx, mod)

	if err != nil {
		return nil, err
	}

	vmctx.modules[name] = lm

	return lm.Get()
}
