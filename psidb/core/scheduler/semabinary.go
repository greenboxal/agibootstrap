package scheduler

import (
	"context"
	"sync"

	"github.com/greenboxal/agibootstrap/psidb/core/api"
)

type BinarySemaphore struct {
	m       sync.Mutex
	sch     *Scheduler
	count   uint64
	done    chan struct{}
	waiters coreapi.ListHead[coreapi.CompletionWaiter]
}

func NewBinarySemaphore(sch *Scheduler, initial bool) *BinarySemaphore {
	bc := &BinarySemaphore{
		sch: sch,
	}

	if initial {
		bc.Reset()
	}

	return bc
}

func (sc *BinarySemaphore) IsComplete() bool {
	return sc.count == 0
}

func (sc *BinarySemaphore) AddWaiterSlot(waiter *coreapi.CompletionWaiterSlot) {
	sc.m.Lock()
	defer sc.m.Unlock()

	if sc.IsComplete() {
		sc.sch.notifyWaiter(sc, waiter.Value)
		return
	}

	sc.waiters.AddPrevious(waiter)
}

func (sc *BinarySemaphore) AddWaiter(waiter coreapi.CompletionWaiter) {
	item := &coreapi.ListHead[coreapi.CompletionWaiter]{Value: waiter}

	sc.AddWaiterSlot(item)
}

func (sc *BinarySemaphore) WaitForCompletion(ctx context.Context) error {
	if sc.done == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case _, _ = <-sc.done:
		return nil
	}
}

func (sc *BinarySemaphore) Acquire(amount uint64) {
	sc.Complete(int64(amount))
}

func (sc *BinarySemaphore) Release(amount uint64) bool {
	return sc.Complete(-int64(amount))
}

func (sc *BinarySemaphore) Complete(amount int64) bool {
	if amount == 0 {
		return sc.count == 0
	}

	sc.m.Lock()
	defer sc.m.Unlock()

	i := int64(sc.count) + amount
	v := uint64(i)

	if i < 0 {
		v = 0
	} else if i > 1 {
		v = 1
	}

	if sc.count == 0 && v > 0 {
		sc.Reset()
	} else if sc.count != 0 && v == 0 {
		if sc.doComplete() {
			sc.scheduleNextWaiter(true)
		}
	}

	sc.count = v

	return sc.count == 0
}

func (sc *BinarySemaphore) Reset() {
	sc.m.Lock()
	defer sc.m.Unlock()

	if sc.done != nil {
		return
	}

	sc.done = make(chan struct{})
}

func (sc *BinarySemaphore) scheduleNextWaiter(recurse bool) bool {
	if waiter := sc.waiters.Next; waiter != nil {
		sc.sch.ScheduleNext(func() {
			sc.sch.notifyWaiter(sc, waiter.Value)

			if recurse {
				sc.scheduleNextWaiter(true)
			}
		})

		waiter.Remove()

		return true
	}

	return false
}

func (sc *BinarySemaphore) doComplete() bool {
	if sc.done == nil {
		return false
	}

	close(sc.done)
	sc.done = nil

	return true
}
