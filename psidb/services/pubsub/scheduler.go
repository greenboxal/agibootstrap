package pubsub

import (
	"sync"
	"time"

	"github.com/jbenet/goprocess"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

type Scheduler struct {
	dispatcher Dispatcher
	tracker    coreapi.ConfirmationTracker
	journal    *graphfs.Journal

	queuesMutex sync.RWMutex
	queues      map[string]*schedulerQueue

	pendingMutex sync.RWMutex
	pending      map[scheduledItemKey]*scheduledItem

	promisesMutex sync.RWMutex
	promises      map[psi.PromiseHandle]*pendingPromise

	proc   goprocess.Process
	closed bool
}

type Dispatcher interface {
	Dispatch(entry *graphfs.JournalEntry)
}

func NewScheduler(journal *graphfs.Journal, tracker coreapi.ConfirmationTracker, dispatcher Dispatcher) *Scheduler {
	sch := &Scheduler{
		tracker:    tracker,
		journal:    journal,
		dispatcher: dispatcher,

		queues:   map[string]*schedulerQueue{},
		pending:  map[scheduledItemKey]*scheduledItem{},
		promises: map[psi.PromiseHandle]*pendingPromise{},
	}

	sch.proc = goprocess.Go(sch.run)

	return sch
}

func (sch *Scheduler) Recover() error {
	sch.pendingMutex.Lock()
	defer sch.pendingMutex.Unlock()

	if sch.closed {
		panic("scheduler closed")
	}

	iter, err := sch.tracker.Recover()

	if err != nil {
		panic(err)
	}

	for iter.Next() {
		ticket := iter.Value()
		entry, err := sch.journal.Read(ticket, nil)

		if err != nil {
			return err
		}

		sch.dispatchUnlocked(entry)
	}

	return nil
}
func (sch *Scheduler) Dispatch(entry *graphfs.JournalEntry) {
	sch.pendingMutex.Lock()
	defer sch.pendingMutex.Unlock()

	if sch.closed {
		panic("scheduler closed")
	}

	sch.dispatchUnlocked(entry)
}

func (sch *Scheduler) Confirm(entry *graphfs.JournalEntry) {
	sch.pendingMutex.Lock()

	entryKey := scheduledItemKey{
		Xid:   entry.Confirmation.Xid,
		Rid:   entry.Confirmation.Rid,
		Nonce: entry.Confirmation.Nonce,
	}

	item := sch.pending[entryKey]

	if item == nil {
		sch.pendingMutex.Unlock()
		return
	}

	item.m.Lock()
	sch.pendingMutex.Unlock()

	defer item.m.Unlock()

	sch.onConfirm(item, entry)
}

func (sch *Scheduler) Wait(entry *graphfs.JournalEntry) {
	for _, handle := range entry.Promises {
		sch.resolvePromise(handle)
	}
}

func (sch *Scheduler) Signal(entry *graphfs.JournalEntry) {
	for _, handle := range entry.Promises {
		sch.resolvePromise(handle)
	}
}

func (sch *Scheduler) Close() error {
	sch.pendingMutex.Lock()
	defer sch.pendingMutex.Unlock()

	sch.closed = true

	return sch.tracker.Close()
}

func (sch *Scheduler) dispatchUnlocked(entry *graphfs.JournalEntry) {
	entryKey := scheduledItemKey{
		Xid:   entry.Xid,
		Rid:   entry.Rid,
		Nonce: entry.Notification.Nonce,
	}

	item := sch.pending[entryKey]

	if item != nil {
		return
	}

	item = &scheduledItem{
		skey:  entryKey,
		qkey:  entry.Notification.Notified.String(),
		entry: entry,
	}

	item.deps = make([]*pendingPromise, 0, len(entry.Notification.Dependencies))

	for _, dep := range entry.Notification.Dependencies {
		p := sch.getOrCreatePromise(dep.PromiseHandle, true, true)
		p.refs++
		p.queues = append(p.queues, item.qkey)
		p.count += dep.Count
		p.m.Unlock()

		item.deps = append(item.deps, p)
	}

	sch.pending[entryKey] = item
	sch.tracker.Track(entryKey.Rid)

	for _, dep := range item.deps {
		dep.queues = append(dep.queues, item.qkey)
	}

	queue := sch.getOrCreateQueue(item.qkey, true, true)
	queue.add(item)
	queue.m.Unlock()

	sch.trySchedule(item.qkey)
}

func (sch *Scheduler) trySchedule(qkey string) {
	queue := sch.getOrCreateQueue(qkey, true, false)

	if queue == nil {
		return
	}

	defer queue.m.Unlock()

	item := queue.head

	if item == nil {
		return
	}

	item.m.Lock()

	if item.scheduled && item.deadline.Before(time.Now()) {
		item.scheduled = false
	}

	if sch.canSchedule(item) {
		sch.dispatchItem(item)
	}

	item.m.Unlock()
}

func (sch *Scheduler) canSchedule(item *scheduledItem) bool {
	if item.scheduled || item.confirmation != nil {
		return false
	}

	for _, dep := range item.deps {
		if !dep.resolved {
			return false
		}
	}

	return true
}

func (sch *Scheduler) dispatchItem(item *scheduledItem) {
	item.scheduled = true
	item.start = time.Now()
	item.deadline = item.start.Add(30 * time.Second)

	time.AfterFunc(300*time.Second, func() {
		if item.confirmation == nil {
			sch.trySchedule(item.qkey)
		}
	})

	sch.dispatcher.Dispatch(item.entry)
}

func (sch *Scheduler) onConfirm(item *scheduledItem, confirm *graphfs.JournalEntry) {
	if item.confirmation != nil {
		return
	}

	item.confirmation = confirm

	if confirm == nil {
		sch.trySchedule(item.qkey)
		return
	}

	for _, dep := range item.deps {
		dep.m.Lock()
		dep.refs--

		for i, qitem := range dep.queues {
			if qitem == item.qkey {
				dep.queues = slices.Delete(dep.queues, i, i+1)
				break
			}
		}

		dep.m.Unlock()
	}

	for _, obs := range item.entry.Notification.Observers {
		sch.resolvePromise(obs)
	}

	sch.pendingMutex.Lock()
	defer sch.pendingMutex.Unlock()

	delete(sch.pending, item.skey)

	queue := sch.getOrCreateQueue(item.qkey, true, false)

	if queue != nil {
		queue.remove(item)
		queue.m.Unlock()

		sch.trySchedule(item.qkey)
	}

	sch.tracker.Confirm(item.skey.Rid)
}

func (sch *Scheduler) getOrCreateQueue(qkey string, lock, create bool) *schedulerQueue {
	if create {
		sch.queuesMutex.Lock()
		defer sch.queuesMutex.Unlock()
	} else {
		sch.queuesMutex.RLock()
		defer sch.queuesMutex.RUnlock()
	}

	if queue := sch.queues[qkey]; queue != nil {
		if lock {
			queue.m.Lock()
		}

		return queue
	}

	if !create {
		return nil
	}

	queue := &schedulerQueue{}

	sch.queues[qkey] = queue

	if lock {
		queue.m.Lock()
	}

	return queue
}

func (sch *Scheduler) getOrCreatePromise(key psi.PromiseHandle, lock, create bool) *pendingPromise {
	if create {
		sch.promisesMutex.Lock()
		defer sch.promisesMutex.Unlock()
	} else {
		sch.promisesMutex.RLock()
		defer sch.promisesMutex.RUnlock()
	}

	if pp := sch.promises[key]; pp != nil {
		if lock {
			pp.m.Lock()
		}

		return pp
	}

	if !create {
		return nil
	}

	pp := &pendingPromise{
		key:  key,
		ch:   make(chan struct{}),
		refs: 1,
	}

	sch.promises[key] = pp

	if lock {
		pp.m.Lock()
	}

	return pp
}

func (sch *Scheduler) resolvePromise(p psi.Promise) {
	pp := sch.getOrCreatePromise(p.PromiseHandle, true, true)
	defer pp.m.Unlock()

	if pp.resolved {
		return
	}

	pp.count += p.Count

	if pp.count != 0 {
		return
	}

	pp.resolved = true
	close(pp.ch)

	for _, qkey := range pp.queues {
		sch.trySchedule(qkey)
	}
}

func (sch *Scheduler) run(proc goprocess.Process) {
	ticker := time.NewTicker(1 * time.Second)

	for !sch.closed {
		select {
		case <-proc.Closing():
			return
		case <-ticker.C:
			if sch.closed {
				return
			}

			if err := sch.tracker.Flush(); err != nil {
				logger.Error(err)
			}
		}
	}
}

type scheduledItemKey struct {
	Xid   uint64
	Rid   uint64
	Nonce uint64
}

type scheduledItem struct {
	m sync.Mutex

	skey scheduledItemKey
	qkey string
	deps []*pendingPromise

	entry        *graphfs.JournalEntry
	confirmation *graphfs.JournalEntry

	start     time.Time
	deadline  time.Time
	scheduled bool

	timeout <-chan time.Time

	next *scheduledItem
}

type pendingPromise struct {
	m      sync.Mutex
	key    psi.PromiseHandle
	refs   int
	queues []string

	ch       chan struct{}
	count    int
	resolved bool
}

type schedulerQueue struct {
	m    sync.Mutex
	head *scheduledItem
}

func (q *schedulerQueue) add(item *scheduledItem) {
	if q.head == nil {
		q.head = item
		return
	}

	tail := q.head

	for tail.next != nil {
		tail = tail.next
	}

	tail.next = item
}

func (q *schedulerQueue) remove(item *scheduledItem) {
	if q.head == item {
		q.head = item.next
		return
	}

	tail := q.head

	for tail.next != nil {
		if tail.next == item {
			tail.next = item.next
			return
		}

		tail = tail.next
	}
}
