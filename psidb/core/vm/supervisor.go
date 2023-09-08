package vm

import (
	"context"
	"sync"

	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"github.com/thejerf/suture/v4"
	"go.uber.org/fx"

	"rogchap.com/v8go"
)

type Supervisor struct {
	iso *v8go.Isolate

	proc goprocess.Process
	sup  *suture.Supervisor

	mu               sync.RWMutex
	isolates         []*Isolate
	nextIsolateIndex int
}

func NewSupervisor(
	lc fx.Lifecycle,
) *Supervisor {
	sup := &Supervisor{
		iso: v8go.NewIsolate(),

		sup: suture.NewSimple("vm/isolates"),
	}

	lc.Append(fx.Hook{
		OnStart: sup.Start,
		OnStop:  sup.Stop,
	})

	return sup
}

func (vmsup *Supervisor) NextIsolate(ctx context.Context) *Isolate {
	v := ctx.Value(ctxKeyIsolate)

	if v != nil {
		return v.(*Isolate)
	}

	vmsup.mu.Lock()
	if len(vmsup.isolates) == 0 {
		vmsup.mu.Unlock()
		return vmsup.NewIsolate()
	}
	defer vmsup.mu.Unlock()

	index := vmsup.nextIsolateIndex

	vmsup.nextIsolateIndex = (vmsup.nextIsolateIndex + 1) % len(vmsup.isolates)

	return vmsup.isolates[index]
}

func (vmsup *Supervisor) AcquireIsolate(ctx context.Context) context.Context {
	v := ctx.Value(ctxKeyIsolate)

	if v == nil {
		v = vmsup.NextIsolate(ctx)
		ctx = context.WithValue(ctx, ctxKeyIsolate, v)
	}

	return ctx
}

func (vmsup *Supervisor) NewIsolate() *Isolate {
	return NewIsolate(vmsup)
}

func (vmsup *Supervisor) Close() error {
	for _, iso := range vmsup.isolates {
		if err := iso.Close(); err != nil {
			return err
		}
	}

	vmsup.iso.Dispose()

	return nil
}

func (vmsup *Supervisor) Start(ctx context.Context) error {
	vmsup.proc = goprocess.Go(func(proc goprocess.Process) {
		ctx := goprocessctx.OnClosingContext(proc)

		vmsup.sup.ServeBackground(ctx)
	})

	return nil
}

func (vmsup *Supervisor) Stop(ctx context.Context) error {
	return vmsup.proc.Close()
}

func (vmsup *Supervisor) notifyIsolateCreated(iso *Isolate) {
	vmsup.mu.Lock()
	defer vmsup.mu.Unlock()

	vmsup.isolates = append(vmsup.isolates, iso)
	iso.supervisorToken = vmsup.sup.Add(iso)
}

func (vmsup *Supervisor) notifyIsolateDestroyed(iso *Isolate) {
	vmsup.mu.Lock()
	defer vmsup.mu.Unlock()

	for i := range vmsup.isolates {
		if vmsup.isolates[i] == iso {
			vmsup.isolates = append(vmsup.isolates[:i], vmsup.isolates[i+1:]...)
			return
		}
	}
}
