package scheduler

import (
	"context"
	"sync"

	"golang.org/x/exp/slices"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type CompletionCondition = coreapi.CompletionCondition
type TaskHandle = coreapi.TaskHandle
type TaskStatus = coreapi.TaskStatus

type Task struct {
	sch        *Scheduler
	dispatcher Dispatcher

	handle TaskHandle
	entry  *coreapi.JournalEntry

	mu sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	waitSemaphores   []CompletionCondition
	signalSemaphores []CompletionCondition

	status TaskStatus
	err    error

	waiter coreapi.ListHead[coreapi.CompletionWaiter]
	queue  coreapi.ListHead[*Task]

	cleanup []func()
}

func NewTask(sch *Scheduler, dispatcher Dispatcher, handle TaskHandle, entry *coreapi.JournalEntry) *Task {
	t := &Task{
		sch:        sch,
		dispatcher: dispatcher,
		handle:     handle,
		entry:      entry,

		status: coreapi.TaskStatusIdle,
	}

	return t
}

func (t *Task) Handle() TaskHandle                  { return t.handle }
func (t *Task) BaseContext() context.Context        { return t.ctx }
func (t *Task) JournalEntry() *coreapi.JournalEntry { return t.entry }

func (t *Task) AddWaitSemaphore(sc coreapi.Semaphore, value uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.waitSemaphores = append(t.waitSemaphores, CompletionCondition{
		Completion: sc,
		Value:      value,
	})
}

func (t *Task) AddSignalSemaphore(sc coreapi.Semaphore, value uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.signalSemaphores = append(t.signalSemaphores, CompletionCondition{
		Completion: sc,
		Value:      value,
	})
}

func (t *Task) CanSchedule() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.status == coreapi.TaskStatusIdle || t.status == coreapi.TaskStatusReady
}

func (t *Task) WaitFor(sc coreapi.Semaphore, value uint64) bool {
	if value == 0 {
		return true
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	return t.waitForUnlocked(sc, value)
}

func (t *Task) waitForUnlocked(sc coreapi.Semaphore, value uint64) bool {
	if value == 0 {
		return true
	}

	idx := slices.IndexFunc(t.waitSemaphores, func(i CompletionCondition) bool {
		return i.Completion == sc && i.Value == value
	})

	if sc.Release(value) {
		if idx != -1 {
			t.waitSemaphores = slices.Delete(t.waitSemaphores, idx, idx+1)
		}

		return true
	} else {
		if idx == -1 {
			t.waitSemaphores = append(t.waitSemaphores, CompletionCondition{
				Completion: sc,
				Value:      value,
			})
		} else {
			t.waitSemaphores[idx].Value = value
		}

		t.waiter.Value = t
		sc.AddWaiterSlot(&t.waiter)

		t.setStatus(coreapi.TaskStatusWaiting)

		return false
	}
}

func (t *Task) TrySchedule() (TaskStatus, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status != coreapi.TaskStatusReady && t.status != coreapi.TaskStatusWaiting && t.status != coreapi.TaskStatusIdle {
		return t.status, false
	}

	for _, sc := range t.waitSemaphores {
		if !t.waitForUnlocked(sc.Completion, sc.Value) {
			return t.status, false
		}
	}

	ctx, cancel := context.WithCancel(t.sch.BaseContext())

	t.ctx = ctx
	t.cancel = cancel

	go t.monitor()

	t.setStatus(coreapi.TaskStatusRunning)

	return t.status, true
}

func (t *Task) OnComplete(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cancel()

	t.err = err
	t.ctx = nil
	t.cancel = nil

	not := t.entry.Notification
	logger.Infow("confirming task", "handle", t.handle, "queue", not.Notified, "iface", not.Interface, "action", not.Action)

	t.setStatus(coreapi.TaskStatusComplete)
}

func (t *Task) OnCompleted(sc coreapi.Semaphore) {
	t.TrySchedule()
}

func (t *Task) Cancel() {
	if t.cancel != nil {
		t.cancel()
	}
}

func (t *Task) Close() error {
	t.Cancel()

	t.waiter.Remove()
	t.queue.Remove()

	return nil
}

func (t *Task) monitor() {
	<-t.ctx.Done()

	t.mu.Lock()
	defer t.mu.Unlock()

	t.OnComplete(t.err)
}

func (t *Task) setStatus(status TaskStatus) {
	if t.status == status {
		return
	}

	t.status = status

	onCompleted := func() {
		for _, fn := range t.cleanup {
			fn()
		}

		t.cleanup = nil

		t.queue.Lock()
		hasNext := t.queue.Next != nil
		t.queue.RemoveUnlocked()

		if hasNext {
			qkey := t.entry.Notification.Notified.String()
			t.sch.scheduleQueue(qkey)
		}
		t.queue.Unlock()
	}

	if t.status == coreapi.TaskStatusReady {
		t.sch.ScheduleTask(t)
	} else if t.status == coreapi.TaskStatusComplete {
		for _, sc := range t.signalSemaphores {
			sc.Completion.Acquire(sc.Value)
		}

		t.signalSemaphores = nil

		onCompleted()
	} else if t.status == coreapi.TaskStatusCanceled {
		onCompleted()
	}
}

func (t *Task) DispatchNow() error {
	if t.status != coreapi.TaskStatusRunning {
		return nil
	}

	return t.dispatcher.DispatchTask(t)
}
