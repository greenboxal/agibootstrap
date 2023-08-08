package pubsub

import (
	"context"
	"sync"

	"github.com/alitto/pond"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
	"github.com/greenboxal/agibootstrap/psidb/services/migrations"
)

type Manager struct {
	mu       sync.RWMutex
	core     coreapi.Core
	roots    map[string]*Topic
	stream   *coreapi.ReplicationStreamProcessor
	migrator *migrations.Manager

	workerPool *pond.WorkerPool
}

func NewManager(
	lc fx.Lifecycle,
	core coreapi.Core,
	migrator *migrations.Manager,
) *Manager {
	m := &Manager{
		core:     core,
		migrator: migrator,
		roots:    map[string]*Topic{},

		workerPool: core.Config().Workers.Build(),
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
	slot, err := pm.core.CreateReplicationSlot(ctx, graphfs.ReplicationSlotOptions{
		Name:       "pubsub",
		Persistent: false,
	})

	if err != nil {
		return err
	}

	pm.stream = coreapi.NewReplicationStream(slot, pm.processReplicationMessage)

	if err := pm.migrator.Migrate(ctx, migrationSet); err != nil {
		return err
	}

	return pm.LoadPersistentState(ctx)
}

func (pm *Manager) LoadPersistentState(ctx context.Context) error {
	/*return pm.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
		root, err := tx.Resolve(ctx, RootPath)

		if err != nil {
			return err
		}

		for edges := root.Edges(); edges.Next(); {
			edge := edges.Value()
			to, err := edge.ResolveTo(ctx)

			if err != nil {
				return err
			}

			ps, ok := to.(*PersistentSubscription)

			if !ok {
				continue
			}


		}

		return nil
	})*/

	return nil
}

func (pm *Manager) processReplicationMessage(ctx context.Context, entries []*graphfs.JournalEntry) error {
	var wg sync.WaitGroup

	notifications := make(map[string][]*psi.Notification)

	for _, entry := range entries {
		if entry.Op == graphfs.JournalOpNotify {
			key := entry.Notification.Notified.String()

			notifications[key] = append(notifications[key], entry.Notification)

			continue
		}

		if entry.Path == nil {
			continue
		}

		wg.Add(1)

		go func(entry *graphfs.JournalEntry) {
			defer wg.Done()

			n := Notification{
				Ts:   entry.Ts,
				Path: *entry.Path,
			}

			root := pm.getOrCreateRoot(entry.Path.Root(), false)

			if root == nil {
				return
			}

			root.Push(n)

			for _, el := range entry.Path.Components() {
				root = root.Topic(el.String(), false)

				if root == nil {
					return
				}

				root.Push(n)
			}
		}(entry)
	}

	wg.Wait()

	if len(notifications) > 0 {
		if err := pm.dispatchNotify(ctx, notifications); err != nil {
			return err
		}
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
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, t := range pm.roots {
		if err := t.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (pm *Manager) dispatchNotify(ctx context.Context, notifications map[string][]*psi.Notification) error {
	var merr error

	wg, gctx := pm.workerPool.GroupContext(ctx)

	for _, notifications := range notifications {
		if len(notifications) == 0 {
			continue
		}

		notifications := notifications
		targetPath := notifications[0].Notified

		wg.Submit(func() error {
			return pm.core.RunTransaction(gctx, func(ctx context.Context, tx coreapi.Transaction) error {
				target, err := tx.Resolve(ctx, targetPath)

				if err != nil {
					return err
				}

				for _, not := range notifications {
					if err := not.Apply(ctx, target); err != nil {
						merr = multierror.Append(merr, err)
					}
				}

				return nil
			})
		})
	}

	if err := wg.Wait(); err != nil {
		return nil
	}

	return merr
}
