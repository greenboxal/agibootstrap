package pubsub

import (
	"context"
	"sync"

	"github.com/jbenet/goprocess"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Topic struct {
	mu sync.RWMutex

	name     string
	path     psi.Path
	parent   *Topic
	children map[string]*Topic

	queue chan Notification
	proc  goprocess.Process

	subscriptions map[string]*Subscription
}

func NewTopic(parent *Topic, path psi.Path) *Topic {
	t := &Topic{
		path:          path,
		parent:        parent,
		children:      make(map[string]*Topic),
		subscriptions: make(map[string]*Subscription),
		queue:         make(chan Notification, 256),
	}

	if path.Len() > 0 {
		t.name = path.Components()[path.Len()-1].String()
	}

	t.proc = goprocess.Go(t.run)

	return t
}

func (t *Topic) Publish(ctx context.Context, msg Notification) error {
	select {
	case t.queue <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *Topic) GetChild(name string, create bool) *Topic {
	t.mu.RLock()
	if child, ok := t.children[name]; ok {
		t.mu.RUnlock()
		return child
	}
	t.mu.RUnlock()

	if !create {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if child, ok := t.children[name]; ok {
		return child
	}

	child := NewTopic(t, t.path.Join(psi.MustParsePath(name)))

	t.children[name] = child

	return child
}

func (t *Topic) run(proc goprocess.Process) {
	proc.SetTeardown(t.teardown)

	for {
		select {
		case <-proc.Closing():
			return
		case msg := <-t.queue:
			t.dispatch(msg)
		}
	}
}

func (t *Topic) dispatch(msg Notification) {
	if !msg.Path.Equals(t.path) {
		rel, err := msg.Path.RelativeTo(t.path)

		if err != nil {
			return
		}

		if rel.Len() > 0 {
			child := t.GetChild(rel.Components()[0].String(), false)

			if child != nil {
				child.dispatch(msg)
			}
		}
	}

	for _, subs := range t.subscriptions {
		if subs.IsCompatibleWith(msg) {
			subs.dispatch(msg)
		}
	}
}

func (t *Topic) Close() error {
	return t.proc.Close()
}

func (t *Topic) teardown() error {
	for _, child := range t.children {
		if err := child.Close(); err != nil {
			return err
		}
	}

	if t.parent != nil {
		t.parent.notifyChildClosed(t)
	}

	return nil
}

func (t *Topic) notifyChildClosed(child *Topic) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.children, child.name)
}

func (t *Topic) addSubscription(s *Subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t.mu.Lock()
	defer t.mu.Unlock()

	if s.closed {
		panic("subscription is closed")
	}

	if s.topic != nil && s.topic != t {
		panic("subscription already belongs to another topic")
	}

	s.topic = t
	t.subscriptions[s.id] = s
}

func (t *Topic) removeSubscription(s *Subscription) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.subscriptions, s.id)
}
