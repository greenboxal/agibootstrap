package pubsub

import (
	"sync"
	"time"

	"github.com/jbenet/goprocess"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

type scheduledItemKey struct {
	Xid   uint64
	Rid   uint64
	Nonce uint64
}

type scheduledItem struct {
	m sync.Mutex

	skey scheduledItemKey
	qkey string

	entry        *graphfs.JournalEntry
	confirmation *graphfs.JournalEntry

	start    time.Time
	deadline time.Time

	timeout <-chan time.Time

	next *scheduledItem
}

type Scheduler struct {
	mu sync.RWMutex

	dispatcher Dispatcher
	tracker    coreapi.ConfirmationTracker
	journal    *graphfs.Journal

	queues  map[string]*scheduledItem
	pending map[scheduledItemKey]*scheduledItem

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

		queues:  map[string]*scheduledItem{},
		pending: map[scheduledItemKey]*scheduledItem{},
	}

	sch.proc = goprocess.Go(sch.run)

	return sch
}

func (sch *Scheduler) Recover() error {
	sch.mu.Lock()
	defer sch.mu.Unlock()

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
	sch.mu.Lock()
	defer sch.mu.Unlock()

	sch.dispatchUnlocked(entry)
}

func (sch *Scheduler) dispatchUnlocked(entry *graphfs.JournalEntry) {
	if sch.closed {
		panic("scheduler closed")
	}

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

	item.m.Lock()
	defer item.m.Unlock()

	sch.pending[entryKey] = item
	sch.tracker.Track(entryKey.Rid)

	if queue := sch.queues[item.qkey]; queue != nil {
		queue.m.Lock()

		for queue.next != nil {
			next := queue.next

			next.m.Lock()
			queue.m.Unlock()

			queue = next
		}

		queue.next = item

		queue.m.Unlock()
	} else {
		sch.queues[item.qkey] = item

		sch.reschedule(item)
	}
}

func (sch *Scheduler) Confirm(entry *graphfs.JournalEntry) {
	sch.mu.Lock()
	defer sch.mu.Unlock()

	entryKey := scheduledItemKey{
		Xid:   entry.Confirmation.Xid,
		Rid:   entry.Confirmation.Rid,
		Nonce: entry.Confirmation.Nonce,
	}

	item := sch.pending[entryKey]

	if item == nil {
		return
	}

	item.m.Lock()
	defer item.m.Unlock()

	sch.onConfirm(item, entry)
}

func (sch *Scheduler) Close() error {
	sch.mu.Lock()
	defer sch.mu.Unlock()

	sch.closed = true

	return sch.tracker.Close()
}

func (sch *Scheduler) reschedule(item *scheduledItem) {
	item.start = time.Now()
	item.deadline = item.start.Add(30 * time.Second)

	time.AfterFunc(300*time.Second, func() {
		item.m.Lock()
		defer item.m.Unlock()

		if item.confirmation == nil {
			sch.reschedule(item)
		}
	})

	sch.dispatcher.Dispatch(item.entry)
}

func (sch *Scheduler) onConfirm(item *scheduledItem, confirm *graphfs.JournalEntry) {
	if item.confirmation != nil {
		return
	}

	item.confirmation = confirm

	if item.next != nil {
		sch.queues[item.qkey] = item.next
		sch.reschedule(item.next)
	} else if sch.queues[item.qkey] == item {
		delete(sch.queues, item.qkey)
	}

	delete(sch.pending, item.skey)
	sch.tracker.Confirm(item.skey.Rid)
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
