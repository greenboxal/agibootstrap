package pubsub

import (
	"context"
	"sync"

	"go.uber.org/fx"

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

func (m *Manager) Subscribe(pattern SubscriptionPattern, handler func(notification Notification)) *Subscription {
	root := m.getOrCreateRoot(pattern.Path.Root(), true)

	for _, el := range pattern.Path.Components() {
		root = root.Topic(el.String(), true)
	}

	return root.Subscribe(pattern, handler)
}

func (m *Manager) Start(ctx context.Context) error {
	slot, err := m.core.CreateReplicationSlot(ctx, graphfs.ReplicationSlotOptions{
		Name:       "pubsub",
		Persistent: false,
	})

	if err != nil {
		return err
	}

	m.stream = coreapi.NewReplicationStream(slot, m.processReplicationMessage)

	if err := m.migrator.Migrate(ctx, migrationSet); err != nil {
		return err
	}

	return m.LoadPersistentState(ctx)
}

func (m *Manager) LoadPersistentState(ctx context.Context) error {
	/*return m.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
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

func (m *Manager) processReplicationMessage(ctx context.Context, entries []*graphfs.JournalEntry) error {
	var wg sync.WaitGroup

	for _, entry := range entries {
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

			root := m.getOrCreateRoot(entry.Path.Root(), false)

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

	return nil
}

func (m *Manager) getOrCreateRoot(name string, create bool) *Topic {
	if name == "" {
		panic("empty topic name")
	}

	if !create {
		m.mu.RLock()
		defer m.mu.RUnlock()

		return m.roots[name]
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if t := m.roots[name]; t != nil {
		return t
	}

	t := NewTopic(m, name)

	m.roots[t.name] = t

	return t
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, t := range m.roots {
		if err := t.Close(); err != nil {
			return err
		}
	}

	return nil
}
