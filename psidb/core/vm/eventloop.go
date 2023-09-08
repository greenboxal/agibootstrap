package vm

import (
	"context"

	"github.com/jbenet/goprocess"
)

type RunQueueItem func(ctx context.Context)
type RunQueuePanicHandler func(any)

type RunQueue struct {
	queue chan RunQueueItem
	proc  goprocess.Process
	done  chan struct{}

	panicHandler RunQueuePanicHandler
	baseContext  context.Context
}

func NewRunQueue(baseContext context.Context, panicHandler RunQueuePanicHandler) *RunQueue {
	ev := &RunQueue{
		queue:        make(chan RunQueueItem, 1024),
		panicHandler: panicHandler,
		baseContext:  baseContext,
	}

	return ev
}

func (ev *RunQueue) UseContext(ctx context.Context) {
	ev.baseContext = ctx
}

func (ev *RunQueue) Done() <-chan struct{} {
	return ev.proc.Closed()
}

func (ev *RunQueue) Run(proc goprocess.Process) {
	ev.proc = proc

	defer close(ev.queue)

	for {
		select {
		case item, ok := <-ev.queue:
			if !ok {
				return
			}

			func() {
				if ev.panicHandler != nil {
					defer func() {
						if r := recover(); r != nil {
							ev.panicHandler(r)
						}
					}()
				}

				item(ev.baseContext)
			}()

		case <-proc.Closing():
			return
		}
	}
}

func (ev *RunQueue) Dispatch(item RunQueueItem) {
	ev.queue <- item
}

func (ev *RunQueue) Close() error {
	return ev.proc.Close()
}
