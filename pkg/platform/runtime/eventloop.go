package runtime

import (
	"sync"
	"sync/atomic"

	"github.com/jbenet/goprocess"

	"github.com/greenboxal/agibootstrap/pkg/platform"
)

type eventLoop struct {
	mu   sync.Mutex
	cond *sync.Cond
	proc goprocess.Process

	ch    chan func()
	queue []func()

	nestingLevel atomic.Int64
}

func NewEventLoop() platform.EventLoop {
	el := &eventLoop{}

	el.cond = sync.NewCond(&el.mu)
	el.proc = goprocess.Go(el.run)

	return el
}

func (e *eventLoop) Dispatch(f func()) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.queue = append(e.queue, f)
	e.cond.Signal()
}

func (e *eventLoop) dequeue(wait bool) (func(), bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for len(e.queue) == 0 {
		if !wait {
			return nil, false
		}
		e.cond.Wait()
	}

	f := e.queue[0]
	e.queue = e.queue[1:]

	return f, true
}

func (e *eventLoop) EnterNestedEventLoop(wait bool) {
	level := e.nestingLevel.Add(1)

	for e.nestingLevel.Load() >= level {
		f, ok := e.dequeue(wait)

		if !ok {
			return
		}

		e.ch <- f
	}
}

func (e *eventLoop) ExitNestedEventLoop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.nestingLevel.Load() == 0 {
		panic("attempted to exit event loop when not in one")
	}

	e.nestingLevel.Add(-1)
}

func (e *eventLoop) Close() error {
	if e.proc == nil {
		return nil
	}

	return e.proc.Close()
}

func (e *eventLoop) run(proc goprocess.Process) {
	defer func() {
		if err := proc.CloseAfterChildren(); err != nil {
			panic(err)
		}
	}()

	for {
		select {
		case <-proc.Closing():
			return

		case f := <-e.ch:
			func() {
				defer func() {
					if e := recover(); e != nil {
						// TODO: log
						panic(e)
					}
				}()

				f()
			}()
		}
	}
}
