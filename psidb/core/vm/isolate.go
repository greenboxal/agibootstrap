package vm

import (
	"context"
	"sync"

	"github.com/jbenet/goprocess"
	"github.com/thejerf/suture/v4"

	"rogchap.com/v8go"
)

type Isolate struct {
	m sync.RWMutex

	supervisor      *Supervisor
	supervisorToken suture.ServiceToken

	iso   *v8go.Isolate
	queue *RunQueue

	contextMap map[*v8go.Context]*Context

	prototypeCache *RuntimePrototypeCache
	moduleCache    *ModuleCache
}

var ctxKeyIsolate = &struct{ ctxKeyIsolate string }{"ctxKeyIsolate"}

func NewIsolate(sup *Supervisor) *Isolate {
	iso := &Isolate{
		supervisor: sup,
		iso:        v8go.NewIsolate(),

		contextMap: map[*v8go.Context]*Context{},
	}

	iso.moduleCache = NewModuleCache(sup)
	iso.prototypeCache = NewRuntimePrototypeCache(iso)
	iso.queue = NewRunQueue(context.WithValue(context.Background(), ctxKeyIsolate, iso), iso.handlePanic)

	sup.notifyIsolateCreated(iso)

	return iso
}

func (iso *Isolate) Isolate() *v8go.Isolate { return iso.iso }

func (iso *Isolate) Close() error {
	if err := iso.queue.Close(); err != nil {
		return err
	}

	iso.iso.Dispose()

	iso.supervisor.notifyIsolateDestroyed(iso)

	return nil
}

func (iso *Isolate) registerContext(vmctx *Context) {
	iso.m.Lock()
	defer iso.m.Unlock()

	iso.contextMap[vmctx.ctx] = vmctx
}

func (iso *Isolate) unregisterContext(vmctx *Context) {
	iso.m.Lock()
	defer iso.m.Unlock()

	delete(iso.contextMap, vmctx.ctx)
}

func (iso *Isolate) getContext(context *v8go.Context) *Context {
	iso.m.RLock()
	defer iso.m.RUnlock()

	return iso.contextMap[context]
}

func (iso *Isolate) handlePanic(a any) {
	logger.Error(a)
}

func (iso *Isolate) Serve(ctx context.Context) error {
	iso.queue.UseContext(context.WithValue(ctx, ctxKeyIsolate, iso))

	logger.Infow("Starting isolate")

	proc := goprocess.Go(iso.queue.Run)

	select {
	case <-ctx.Done():
	case <-proc.Closed():
	}

	return proc.Close()
}
