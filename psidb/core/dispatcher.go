package core

import (
	"context"
	"sync"

	"github.com/go-errors/errors"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	scheduler "github.com/greenboxal/agibootstrap/psidb/core/scheduler"
)

type Dispatcher struct {
	core           coreapi.Core
	scheduler      *scheduler.Scheduler
	syncManager    *scheduler.SyncManager
	sessionManager coreapi.SessionManager

	tracker coreapi.ConfirmationTracker
	stream  *coreapi.ReplicationStreamProcessor

	mu           sync.RWMutex
	pendingTasks map[coreapi.TaskHandle]*scheduler.Task
}

func NewDispatcher(
	lc fx.Lifecycle,
	core coreapi.Core,
	sch *scheduler.Scheduler,
	syncm *scheduler.SyncManager,
	sm coreapi.SessionManager,
) *Dispatcher {
	d := &Dispatcher{
		core:           core,
		scheduler:      sch,
		syncManager:    syncm,
		sessionManager: sm,

		pendingTasks: map[coreapi.TaskHandle]*scheduler.Task{},
	}

	lc.Append(fx.Hook{
		OnStart: d.OnStart,
		OnStop:  d.OnStop,
	})

	return d
}

func (d *Dispatcher) AcceptEntry(entry *coreapi.JournalEntry) {
	switch entry.Op {
	case coreapi.JournalOpNotify:
		d.acceptNotify(entry)

	case coreapi.JournalOpConfirm:
		d.acceptConfirm(entry)

	case coreapi.JournalOpWait:
		for _, handle := range entry.Promises {
			sema := d.syncManager.GetOrCreateSemaphore(handle.PromiseHandle)
			sema.Release(uint64(handle.Count))
		}

	case coreapi.JournalOpSignal:
		for _, handle := range entry.Promises {
			sema := d.syncManager.GetOrCreateSemaphore(handle.PromiseHandle)
			sema.Acquire(uint64(handle.Count))
		}
	}
}

func (d *Dispatcher) acceptNotify(entry *coreapi.JournalEntry) {
	d.mu.Lock()
	defer d.mu.Unlock()

	handle := coreapi.TaskHandle{
		Xid:   entry.Xid,
		Rid:   entry.Rid,
		Nonce: entry.Notification.Nonce,
	}

	if _, ok := d.pendingTasks[handle]; ok {
		return
	}

	task := scheduler.NewTask(d.scheduler, d, handle, entry)

	d.pendingTasks[task.Handle()] = task
	d.tracker.Track(handle.Rid)
	d.scheduler.ScheduleTask(task)
}

func (d *Dispatcher) acceptConfirm(entry *coreapi.JournalEntry) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.tracker.Confirm(entry.Confirmation.Rid)

	handle := coreapi.TaskHandle{
		Xid:   entry.Confirmation.Xid,
		Rid:   entry.Confirmation.Rid,
		Nonce: entry.Confirmation.Nonce,
	}

	task := d.pendingTasks[handle]

	if task == nil {
		return
	}

	var err error

	if !entry.Confirmation.Ok {
		err = errors.New("confirmation failed")
	}

	task.OnComplete(err)

	delete(d.pendingTasks, handle)
}

func (d *Dispatcher) DispatchTask(task *scheduler.Task) error {
	entry, err := d.core.Journal().Read(task.Handle().Rid, nil)

	if err != nil {
		return err
	}

	not := entry.Notification

	ctx, cancel := context.WithCancel(task.BaseContext())
	defer cancel()

	ctx, span := tracer.Start(ctx, "Scheduler.Dispatch")
	span.SetAttributes(semconv.ServiceName("NodeRunner"))
	span.SetAttributes(semconv.RPCSystemKey.String("psidb-node"))
	span.SetAttributes(semconv.RPCService(not.Interface))
	span.SetAttributes(semconv.RPCMethod(not.Action))

	logger := logging.GetLoggerCtx(ctx, "psidb-dispatcher")

	defer func() {
		if e := recover(); e != nil {
			span.SetStatus(codes.Error, "panic")
			span.RecordError(errors.Wrap(e, 1))

			logger.Errorw("panic", "error", e)
		}

		span.End()
	}()

	if not.SessionID != "" {
		sess := coreapi.GetSession(ctx)

		if sess == nil || sess.UUID() != not.SessionID {
			sess = d.sessionManager.GetOrCreateSession(coreapi.SessionConfig{
				SessionID: not.SessionID,
			})

			ctx = coreapi.WithSession(ctx, sess)
		}
	}

	err = d.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
		target, err := tx.Resolve(ctx, not.Notified)

		if err != nil {
			return err
		}

		ack := psi.Confirmation{
			Xid:   entry.Xid,
			Rid:   entry.Rid,
			Nonce: not.Nonce,
		}

		if _, err := not.Apply(ctx, target); err != nil {
			logger.Errorw("error applying notification", "error", err)

			span.RecordError(err)

			ack.Ok = false
		} else {
			ack.Ok = true
		}

		if ack.Ok {
			span.SetStatus(codes.Ok, "")
		}

		return tx.Confirm(ctx, ack)
	})

	if err != nil {
		logger.Errorw("error running transaction", "error", err)

		span.SetStatus(codes.Error, "error running transaction")
		span.RecordError(err)
	}

	return nil
}

func (d *Dispatcher) OnStart(ctx context.Context) error {
	t, err := d.core.CreateConfirmationTracker(ctx, "core-dispatcher")

	if err != nil {
		return err
	}

	d.tracker = t

	slot, err := d.core.CreateReplicationSlot(ctx, coreapi.ReplicationSlotOptions{
		Name:       "core-dispatcher",
		Persistent: true,
	})

	if err != nil {
		return err
	}

	if err := d.Recover(); err != nil {
		return err
	}

	d.stream = coreapi.NewReplicationStream(slot, d.processReplicationMessage)

	return nil
}

func (d *Dispatcher) processReplicationMessage(ctx context.Context, entry []*coreapi.JournalEntry) error {
	for _, e := range entry {
		d.AcceptEntry(e)
	}

	return nil
}

func (d *Dispatcher) Recover() error {
	iter, err := d.tracker.Recover()

	if err != nil {
		return err
	}

	for iter.Next() {
		ticket := iter.Value()
		entry, err := d.core.Journal().Read(ticket, nil)

		if err != nil {
			return err
		}

		d.acceptNotify(entry)
	}

	return nil
}

func (d *Dispatcher) OnStop(ctx context.Context) error {
	if d.stream != nil {
		if err := d.stream.Close(ctx); err != nil {
			return err
		}
	}

	if d.tracker != nil {
		if err := d.tracker.Close(); err != nil {
			return err
		}
	}

	return nil
}
