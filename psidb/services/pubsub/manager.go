package pubsub

import (
	"context"
	"sync"

	"github.com/alitto/pond"
	"github.com/go-errors/errors"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
	"github.com/greenboxal/agibootstrap/psidb/services/migrations"
)

var logger = logging.GetLogger("pubsub")

type Manager struct {
	mu    sync.RWMutex
	roots map[string]*Topic

	core           coreapi.Core
	sessionManager coreapi.SessionManager

	stream   *coreapi.ReplicationStreamProcessor
	migrator migrations.Migrator

	scheduler  *OldScheduler
	workerPool *pond.WorkerPool

	rootCtx       context.Context
	rootCtxCancel context.CancelFunc
}

func NewManager(
	lc fx.Lifecycle,
	core coreapi.Core,
	sm coreapi.SessionManager,
	migrator migrations.Migrator,
) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		core:           core,
		migrator:       migrator,
		sessionManager: sm,

		roots:      map[string]*Topic{},
		workerPool: core.Config().Workers.Build(),

		rootCtx:       ctx,
		rootCtxCancel: cancel,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return m.Start(ctx)
		},

		OnStop: func(ctx context.Context) error {
			return m.Close()
		},
	})

	return m
}

func (pm *Manager) Scheduler() *OldScheduler { return pm.scheduler }

func (pm *Manager) Subscribe(pattern SubscriptionPattern, handler func(notification Notification)) *Subscription {
	root := pm.getOrCreateRoot(pattern.Path.Root(), true)

	for _, el := range pattern.Path.Components() {
		root = root.Topic(el.String(), true)
	}

	return root.Subscribe(pattern, handler)
}

func (pm *Manager) Start(ctx context.Context) error {
	tracker, err := pm.core.CreateConfirmationTracker(ctx, "pubsub")

	if err != nil {
		return err
	}

	slot, err := pm.core.CreateReplicationSlot(ctx, graphfs.ReplicationSlotOptions{
		Name:       "pubsub",
		Persistent: false,
	})

	if err != nil {
		return err
	}

	if err := pm.migrator.Migrate(ctx, migrationSet); err != nil {
		return err
	}

	pm.scheduler = NewScheduler(pm.core.Journal(), tracker, pm)

	if err := pm.scheduler.Recover(); err != nil {
		return err
	}

	pm.stream = coreapi.NewReplicationStream(slot, pm.processReplicationMessage)

	return nil
}

func (pm *Manager) processReplicationMessage(ctx context.Context, entries []*graphfs.JournalEntry) error {
	wg, _ := pm.workerPool.GroupContext(ctx)

	for _, entry := range entries {
		if entry.Op == graphfs.JournalOpNotify && entry.Confirmation == nil {
			pm.scheduler.Dispatch(entry)
		} else if entry.Op == graphfs.JournalOpConfirm {
			pm.scheduler.Confirm(entry)
		} else if entry.Op == graphfs.JournalOpWait {
			pm.scheduler.Wait(entry)
		} else if entry.Op == graphfs.JournalOpSignal {
			pm.scheduler.Signal(entry)
		}
	}

	for _, entry := range entries {
		entry := entry

		if entry.Path == nil {
			continue
		}

		wg.Submit(func() error {
			n := Notification{
				Ts:   entry.Ts,
				Path: *entry.Path,
			}

			root := pm.getOrCreateRoot(entry.Path.Root(), false)

			if root == nil {
				return nil
			}

			root.Push(n)

			for _, el := range entry.Path.Components() {
				root = root.Topic(el.String(), false)

				if root == nil {
					return nil
				}

				root.Push(n)
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return err
	}

	return pm.scheduler.tracker.Flush()
}

func (pm *Manager) getOrCreateRoot(name string, create bool) *Topic {
	if name == "" {
		panic("empty topic name")
	}

	if !create {
		pm.mu.RLock()
		defer pm.mu.RUnlock()

		return pm.roots[name]
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if t := pm.roots[name]; t != nil {
		return t
	}

	t := NewTopic(pm, name)

	pm.roots[t.name] = t

	return t
}

func (pm *Manager) Close() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.rootCtxCancel()

	if err := pm.scheduler.Close(); err != nil {
		return err
	}

	for _, t := range pm.roots {
		if err := t.Close(); err != nil {
			return err
		}
	}

	pm.workerPool.Stop()

	return nil
}

func (pm *Manager) Dispatch(entry *graphfs.JournalEntry) {
	pm.workerPool.TrySubmit(func() {
		not := entry.Notification

		ctx, cancel := context.WithCancel(pm.rootCtx)
		defer cancel()

		ctx, span := tracer.Start(ctx, "Scheduler.Dispatch")
		span.SetAttributes(semconv.ServiceName("NodeRunner"))
		span.SetAttributes(semconv.RPCSystemKey.String("psidb-node"))
		span.SetAttributes(semconv.RPCService(not.Interface))
		span.SetAttributes(semconv.RPCMethod(not.Action))

		defer func() {
			if e := recover(); e != nil {
				span.SetStatus(codes.Error, "panic")
				span.RecordError(errors.Wrap(e, 1))

				logger.Error(e)
			}

			span.End()
		}()

		if not.SessionID != "" {
			sess := coreapi.GetSession(ctx)

			if sess == nil || sess.UUID() != not.SessionID {
				sess = pm.sessionManager.GetOrCreateSession(not.SessionID)

				ctx = coreapi.WithSession(ctx, sess)
			}
		}

		err := pm.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
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
				logger.Error(err)

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
			logger.Error(err)

			span.SetStatus(codes.Error, "error running transaction")
			span.RecordError(err)
		}
	})
}
