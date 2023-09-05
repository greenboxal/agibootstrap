package pubsub

import (
	"context"
	"sync"

	"github.com/alitto/pond"
	"go.uber.org/fx"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/services/migrations"
)

type Manager struct {
	mu    sync.RWMutex
	roots map[string]*Topic

	core           coreapi.Core
	sessionManager coreapi.SessionManager

	stream   *coreapi.ReplicationStreamProcessor
	migrator migrations.Migrator

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

func (pm *Manager) Subscribe(pattern SubscriptionPattern, handler func(notification Notification)) *Subscription {
	root := pm.getOrCreateRoot(pattern.Path.Root(), true)

	for _, el := range pattern.Path.Components() {
		root = root.Topic(el.String(), true)
	}

	return root.Subscribe(pattern, handler)
}

func (pm *Manager) Start(ctx context.Context) error {
	slot, err := pm.core.CreateReplicationSlot(ctx, coreapi.ReplicationSlotOptions{
		Name:       "pubsub",
		Persistent: false,
	})

	if err != nil {
		return err
	}

	if err := pm.migrator.Migrate(ctx, migrationSet); err != nil {
		return err
	}

	pm.stream = coreapi.NewReplicationStream(slot, pm.processReplicationMessage)

	return nil
}

func (pm *Manager) processReplicationMessage(ctx context.Context, entries []*coreapi.JournalEntry) error {
	wg, _ := pm.workerPool.GroupContext(ctx)

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

	return nil
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
	pm.rootCtxCancel()

	for _, t := range pm.roots {
		if err := t.Close(); err != nil {
			return err
		}
	}

	pm.workerPool.Stop()

	return nil
}
