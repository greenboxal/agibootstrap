package fti

import (
	"context"

	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
)

type taskContext struct {
	ctx            context.Context
	proc           goprocess.Process
	current, total int
	err            error
	done           chan struct{}
}

func (t *taskContext) Context() context.Context {
	return t.ctx
}

func (t *taskContext) Update(current, total int) {
	t.current = current
	t.total = total
}

func (t *taskContext) Err() error {
	t.Wait()
	return t.err
}

func (t *taskContext) Cancel() {

}

func (t *taskContext) Wait() {
	_, _ = <-t.done
}

func SpawnTask(ctx context.Context, task Task) TaskHandle {
	tc := &taskContext{ctx: ctx}

	parent := goprocessctx.WithContext(ctx)

	tc.proc = goprocess.GoChild(parent, func(proc goprocess.Process) {
		defer proc.CloseAfterChildren()
		defer close(tc.done)

		if err := task.Run(tc); err != nil {
			tc.err = err
		}
	})

	return tc
}

type TaskProgress interface {
	Context() context.Context
	Update(current, total int)
}

type TaskHandle interface {
	Context() context.Context

	Cancel()
	Wait()
	Err() error
}

type TaskFunc func(progress TaskProgress) error

func (f TaskFunc) Run(progress TaskProgress) error {
	return f(progress)
}

type Task interface {
	Run(progress TaskProgress) error
}
