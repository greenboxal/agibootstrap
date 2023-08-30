package scheduler

import (
	"context"
	"sync"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type CountingSemaphore struct {
	m       sync.RWMutex
	cond    sync.Cond
	sch     *Scheduler
	count   uint64
	waiters coreapi.CompletionWaiterSlot
}

func NewCountingSemaphore(sch *Scheduler, initial uint64) *CountingSemaphore {
	return &CountingSemaphore{
		sch:   sch,
		count: initial,
	}
}

func (sc *CountingSemaphore) IsComplete() bool {
	sc.m.RLock()
	defer sc.m.RUnlock()

	return sc.count == 0
}
func (sc *CountingSemaphore) AddWaiter(waiter coreapi.CompletionWaiter) {
	item := &coreapi.ListHead[coreapi.CompletionWaiter]{Value: waiter}

	sc.AddWaiterSlot(item)
}

func (sc *CountingSemaphore) AddWaiterSlot(waiter *coreapi.CompletionWaiterSlot) {
	sc.m.Lock()
	defer sc.m.Unlock()

	if sc.IsComplete() {
		sc.sch.notifyWaiter(sc, waiter.Value)
		return
	}

	sc.waiters.AddPrevious(waiter)
}

func (sc *CountingSemaphore) WaitForCompletion(ctx context.Context) error {
	ch := make(chan struct{})

	sc.AddWaiter(coreapi.CompletionWaiterFunc(func(_ coreapi.Semaphore) {
		close(ch)
	}))

	select {
	case <-ctx.Done():
		return ctx.Err()
	case _, _ = <-ch:
		return nil
	}
}

func (sc *CountingSemaphore) Acquire(amount uint64) {
	sc.m.Lock()
	defer sc.m.Unlock()

	sc.count += amount
}

func (sc *CountingSemaphore) Release(amount uint64) bool {
	sc.m.Lock()
	defer sc.m.Unlock()

	if sc.count < amount {
		return false
	}

	sc.count -= amount

	if sc.count > 0 {
		return false
	}

	sc.count = 0

	sc.scheduleNextWaiter(true)

	return true
}

func (sc *CountingSemaphore) Reset(initial uint64) {
	sc.m.Lock()
	defer sc.m.Unlock()

	sc.count = initial
}

func (sc *CountingSemaphore) scheduleNextWaiter(recurse bool) bool {
	if waiter := sc.waiters.Next; waiter != nil {
		sc.sch.ScheduleNext(func() {
			sc.sch.notifyWaiter(sc, waiter.Value)

			if recurse {
				sc.m.Lock()
				defer sc.m.Unlock()

				if sc.count == 0 {
					sc.scheduleNextWaiter(true)
				}
			}
		})

		waiter.Remove()

		return true
	}

	return false
}
