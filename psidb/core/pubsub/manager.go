package pubsub

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type SubscriptionPattern struct {
	ID    string   `json:"id"`
	Path  psi.Path `json:"path"`
	Depth int      `json:"depth"`
}

type Notification struct {
	Ts   int64    `json:"ts"`
	Path psi.Path `json:"path"`
}

type Manager struct {
	root *Topic
}

func NewManager(
	lc fx.Lifecycle,
) *Manager {
	m := &Manager{}
	m.root = NewTopic(nil, psi.MustParsePath("QmYXZ//"))

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return m.Close()
		},
	})

	return m
}

func (pm *Manager) Close() error {
	return pm.root.Close()
}

func (pm *Manager) Publish(ctx context.Context, msg Notification) error {
	return pm.root.Publish(ctx, msg)
}

func (pm *Manager) Subscribe(pattern SubscriptionPattern, handler SubscriptionHandler) *Subscription {
	if pattern.ID == "" {
		pattern.ID = uuid.NewString()
	}

	topic := pm.root

	for _, el := range pattern.Path.Components() {
		topic = topic.GetChild(el.String(), true)
	}

	sub := NewSubscription(pattern.ID, pattern.Path, pattern.Depth, handler)

	topic.addSubscription(sub)

	return sub
}
