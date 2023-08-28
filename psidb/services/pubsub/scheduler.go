package pubsub

import (
	"context"
	"sync"

	"github.com/jbenet/goprocess"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

type SchedulerItem struct {
	m     sync.RWMutex
	queue *SchedulerQueue
	entry *graphfs.JournalEntry

	resolves  []*SchedulerPromise
	blockedBy []*SchedulerPromise

	running  bool
	complete bool
}

func NewSchedulerItem(queue *SchedulerQueue, entry *graphfs.JournalEntry) *SchedulerItem {
	return &SchedulerItem{
		queue: queue,
		entry: entry,
	}
}

func (si *SchedulerItem) IsRunning() bool {
	return si.running
}

func (si *SchedulerItem) IsComplete() bool {
	return si.complete
}

func (si *SchedulerItem) AddSignal(p ...*SchedulerPromise) {
	si.m.Lock()
	defer si.m.Unlock()

	for _, p := range p {
		idx := slices.IndexFunc(si.resolves, func(p2 *SchedulerPromise) bool {
			return p.key == p2.key
		})

		if idx != -1 {
			continue
		}

		si.resolves = append(si.resolves, p.Ref())
	}
}

func (si *SchedulerItem) AddWait(p ...*SchedulerPromise) {
	si.m.Lock()
	defer si.m.Unlock()

	for _, p := range p {
		idx := slices.IndexFunc(si.blockedBy, func(p2 *SchedulerPromise) bool {
			return p.key == p2.key
		})

		if idx != -1 {
			continue
		}

		si.blockedBy = append(si.blockedBy, p.Ref())
	}
}

func (si *SchedulerItem) CanSchedule() bool {
	si.m.RLock()
	defer si.m.RUnlock()

	for _, p := range si.blockedBy {
		if p != nil && !p.HasCompleted() {
			return false
		}
	}

	return true
}

func (si *SchedulerItem) Dispatch() {
	si.m.Lock()
	defer si.m.Unlock()

	if si.running {
		return
	}

	si.running = true
	si.queue.scheduler.scheduleNow(si)
}

func (si *SchedulerItem) WakeUp() {
	si.m.Lock()
	defer si.m.Unlock()

	if si.running {
		return
	}

	for i, p := range si.blockedBy {
		if p == nil {
			continue
		}

		if p.HasCompleted() {
			p.Unref()
			si.blockedBy[i] = nil
		} else {
			return
		}
	}

	si.blockedBy = nil
	si.queue.notifyWakeUp(si)
}

func (si *SchedulerItem) OnComplete() {
	si.m.Lock()
	defer si.m.Unlock()

	si.running = false

	for i, p := range si.resolves {
		if p == nil {
			continue
		}

		p.Resolve()
		p.Unref()

		si.resolves[i] = nil
	}

	si.queue.notifyWakeUp(si)
}

func (si *SchedulerItem) Close() {
	si.m.Lock()
	defer si.m.Unlock()

	for _, p := range si.blockedBy {
		if p == nil {
			continue
		}

		p.Unref()
	}

	for _, p := range si.resolves {
		if p == nil {
			continue
		}

		p.Unref()
	}

	si.blockedBy = nil
}

type SchedulerPromise struct {
	key psi.PromiseHandle

	m         sync.RWMutex
	counter   int
	refs      int
	waiters   []*SchedulerItem
	scheduler *Scheduler

	ch       chan struct{}
	complete bool
}

func NewSchedulerPromise(sch *Scheduler, key psi.PromiseHandle) *SchedulerPromise {
	return &SchedulerPromise{
		key:       key,
		scheduler: sch,
		ch:        make(chan struct{}),
	}
}

func (sp *SchedulerPromise) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case _, _ = <-sp.ch:
		return nil
	}
}

func (sp *SchedulerPromise) Add(count int) {
	sp.m.Lock()
	defer sp.m.Unlock()

	sp.counter += count

	if sp.counter == 0 {
		sp.resolveUnlocked()
	}
}

func (sp *SchedulerPromise) Resolve() {
	sp.m.Lock()
	defer sp.m.Unlock()

	sp.resolveUnlocked()
}

func (sp *SchedulerPromise) resolveUnlocked() {
	if sp.complete {
		return
	}

	sp.complete = true
	sp.counter = 0

	close(sp.ch)

	for _, w := range sp.waiters {
		w.WakeUp()
	}
}

func (sp *SchedulerPromise) Ref() *SchedulerPromise {
	sp.m.Lock()
	defer sp.m.Unlock()

	sp.refs++

	return sp
}

func (sp *SchedulerPromise) Unref() {
	sp.m.Lock()
	defer sp.m.Unlock()

	sp.refs--

	if sp.refs == 0 {
		sp.waiters = nil
	}
}

func (sp *SchedulerPromise) HasCompleted() bool {
	return sp.complete
}

type SchedulerQueue struct {
	scheduler *Scheduler

	dispatchQueue chan *SchedulerItem
	confirmQueue  chan *psi.Confirmation

	pending []*SchedulerItem
}

func NewSchedulerQueue(sch *Scheduler) *SchedulerQueue {
	return &SchedulerQueue{
		scheduler:     sch,
		dispatchQueue: make(chan *SchedulerItem),
		confirmQueue:  make(chan *psi.Confirmation),
	}
}

func (sq *SchedulerQueue) Add(item *SchedulerItem) {
	sq.dispatchQueue <- item
}

func (sq *SchedulerQueue) Run(proc goprocess.Process) {
	for {
		select {
		case <-proc.Closing():
			return

		case item, ok := <-sq.dispatchQueue:
			if !ok {
				return
			}

			idx := slices.IndexFunc(sq.pending, func(item2 *SchedulerItem) bool {
				return item == item2
			})

			if item.IsComplete() {
				if idx != -1 {
					sq.pending = append(sq.pending[:idx], sq.pending[idx+1:]...)
				}
			} else {
				if idx == -1 {
					sq.pending = append(sq.pending, item)
				}
			}

			sq.processQueueHead()

		case confirmation, ok := <-sq.confirmQueue:
			if !ok {
				return
			}

			for _, pending := range sq.pending {
				if pending.entry.Xid == confirmation.Xid && pending.entry.Rid == confirmation.Rid && pending.entry.Notification.Nonce == confirmation.Nonce {
					pending.OnComplete()
				}
			}
		}
	}
}

func (sq *SchedulerQueue) processQueueHead() {
	if len(sq.pending) == 0 {
		return
	}

	for _, item := range sq.pending {
		if !item.CanSchedule() {
			continue
		}

		item.Dispatch()

		return
	}
}

func (sq *SchedulerQueue) Close() {
	close(sq.dispatchQueue)
}

func (sq *SchedulerQueue) notifyWakeUp(si *SchedulerItem) {
	sq.dispatchQueue <- si
}

func (sq *SchedulerQueue) confirmEntry(confirmation *psi.Confirmation) {
	sq.confirmQueue <- confirmation
}

type Scheduler struct {
	dispatcher Dispatcher

	queueMutex sync.RWMutex
	queues     map[string]*SchedulerQueue

	promisesMutex sync.RWMutex
	promises      map[psi.PromiseHandle]*SchedulerPromise
}

func NewScheduler2(dispatcher Dispatcher) *Scheduler {
	return &Scheduler{
		dispatcher: dispatcher,

		queues:   map[string]*SchedulerQueue{},
		promises: map[psi.PromiseHandle]*SchedulerPromise{},
	}
}

func (sch *Scheduler) GetQueue(name string) *SchedulerQueue {
	sch.queueMutex.RLock()
	defer sch.queueMutex.RUnlock()

	return sch.queues[name]
}

func (sch *Scheduler) AddQueue(name string) *SchedulerQueue {
	sch.queueMutex.Lock()
	defer sch.queueMutex.Unlock()

	if q, ok := sch.queues[name]; ok {
		return q
	}

	sq := NewSchedulerQueue(sch)

	sch.queues[name] = sq

	return sq
}

func (sch *Scheduler) GetPromise(key psi.PromiseHandle) *SchedulerPromise {
	sch.promisesMutex.RLock()
	defer sch.promisesMutex.RUnlock()

	return sch.promises[key]
}

func (sch *Scheduler) AddPromise(key psi.PromiseHandle) *SchedulerPromise {
	sch.promisesMutex.Lock()
	defer sch.promisesMutex.Unlock()

	if p, ok := sch.promises[key]; ok {
		return p
	}

	sp := NewSchedulerPromise(sch, key)

	sch.promises[key] = sp

	return sp
}

func (sch *Scheduler) queueForItem(entry *graphfs.JournalEntry, create bool) *SchedulerQueue {
	k := entry.Path.String()

	if !create {
		return sch.GetQueue(k)
	}

	return sch.AddQueue(k)
}

func (sch *Scheduler) Dispatch(entry *graphfs.JournalEntry) {
	q := sch.queueForItem(entry, true)

	q.Add(NewSchedulerItem(q, entry))
}

func (sch *Scheduler) Confirm(entry *graphfs.JournalEntry) {
	if entry.Confirmation != nil {
		q := sch.queueForItem(entry, false)

		if q != nil {
			q.confirmEntry(entry.Confirmation)
		}
	}

	for _, handle := range entry.Promises {
		p := sch.AddPromise(handle.PromiseHandle)

		p.Add(handle.Count)
	}
}

func (sch *Scheduler) scheduleNow(item *SchedulerItem) {
	sch.dispatcher.Dispatch(item.entry)
}
