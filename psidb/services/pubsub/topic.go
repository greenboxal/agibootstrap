package pubsub

import (
	"sync"

	"github.com/jbenet/goprocess"
)

type Topic struct {
	mu            sync.RWMutex
	m             *Manager
	name          string
	subscriptions map[string]*Subscription

	children map[string]*Topic
}

func NewTopic(m *Manager, name string) *Topic {
	return &Topic{
		m:             m,
		name:          name,
		subscriptions: map[string]*Subscription{},
		children:      map[string]*Topic{},
	}
}

func (t *Topic) Name() string { return t.name }

func (t *Topic) Topic(name string, create bool) *Topic {
	if !create {
		t.mu.RLock()
		defer t.mu.RUnlock()

		return t.children[name]
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if c := t.children[name]; c != nil {
		return c
	}

	c := NewTopic(t.m, name)
	t.children[name] = c

	return c
}

func (t *Topic) Push(n Notification) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	l := n.Path.Len()

	for _, s := range t.subscriptions {
		d := l - s.pattern.Path.Len() - s.pattern.Depth

		if d > 0 || s.pattern.Depth == -1 {
			continue
		}

		s.Push(n)
	}
}

func (t *Topic) Subscribe(pattern SubscriptionPattern, handler func(notification Notification)) *Subscription {
	if pattern.ID == "" {
		panic("empty subscription ID")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if s := t.subscriptions[pattern.ID]; s != nil {
		return s
	}

	s := &Subscription{
		topic:   t,
		pattern: pattern,
		handler: handler,
		ch:      make(chan Notification, 16),
	}

	t.subscriptions[s.pattern.ID] = s

	s.proc = goprocess.Go(s.pump)

	return s
}

func (t *Topic) Unsubscribe(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.subscriptions, id)
}

func (t *Topic) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, c := range t.children {
		if err := c.Close(); err != nil {
			return err
		}
	}

	for _, s := range t.subscriptions {
		if err := s.Close(); err != nil {
			return err
		}
	}

	return nil
}
