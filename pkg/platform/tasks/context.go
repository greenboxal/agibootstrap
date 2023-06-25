package tasks

import (
	"context"
	"sync"

	"github.com/jbenet/goprocess"
)

type taskContext struct {
	mu             sync.Mutex
	t              *task
	ctx            context.Context
	cancel         context.CancelFunc
	proc           goprocess.Process
	current, total int
	err            error
	done           chan struct{}
	complete       bool
}

func (t *taskContext) Context() context.Context {
	return t.ctx
}

func (t *taskContext) Update(current, total int) {
	if t.complete {
		return
	}

	t.current = current
	t.total = total

	t.t.Invalidate()
	t.t.Update()
}

func (t *taskContext) Err() error {
	t.Wait()

	return t.err
}

func (t *taskContext) Cancel() {
	if t.complete {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cancel != nil {
		t.cancel()
	}
}

func (t *taskContext) Wait() {
	if t.complete {
		return
	}

	if t.done != nil {
		_, _ = <-t.done
	}
}

func (t *taskContext) onComplete() {
	if t.complete {
		return
	}

	defer close(t.done)

	t.mu.Lock()
	defer t.mu.Unlock()

	t.cancel = nil
	t.ctx = nil
	t.complete = true

	t.t.Update()
}
