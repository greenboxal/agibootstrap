package pubsub

import (
	"context"
	"sync"
	"time"

	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Subscription struct {
	mu   sync.Mutex
	cond sync.Cond

	id      string
	path    psi.Path
	depth   int
	handler SubscriptionHandler

	topic *Topic

	proc   goprocess.Process
	queue  chan Notification
	wake   chan struct{}
	closed bool

	pending        []*Notification
	nextReadIndex  int
	nextWriteIndex int
}

type SubscriptionHandler func(ctx context.Context, msg Notification) error

func NewSubscription(id string, path psi.Path, depth int, handler SubscriptionHandler) *Subscription {
	s := &Subscription{
		id:      id,
		path:    path,
		depth:   depth,
		handler: handler,
		pending: make([]*Notification, 32),
		queue:   make(chan Notification, 16),
	}

	s.cond = sync.Cond{L: &s.mu}
	s.proc = goprocess.Go(s.run)

	return s
}

func (s *Subscription) Pattern() SubscriptionPattern {
	return SubscriptionPattern{
		ID:    s.id,
		Path:  s.path,
		Depth: s.depth,
	}
}

func (s *Subscription) appendPending(item *Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.nextWriteIndex
	s.nextWriteIndex++

	if s.nextWriteIndex >= len(s.pending) {
		s.nextWriteIndex = 0
	}

	if s.nextWriteIndex == s.nextReadIndex {
		s.nextReadIndex++

		if s.nextReadIndex >= len(s.pending) {
			s.nextReadIndex = 0
		}
	}

	s.pending[index] = item

	s.cond.Signal()
}

func (s *Subscription) nextPending(wait bool) *Notification {
	s.mu.Lock()
	defer s.mu.Unlock()

	for len(s.pending) == 0 || s.nextReadIndex == s.nextWriteIndex {
		if !wait {
			s.nextReadIndex = 0
			s.nextWriteIndex = 0
			return nil
		}

		s.cond.Wait()

		if s.closed {
			return nil
		}
	}

	item := s.pending[s.nextReadIndex]
	s.nextReadIndex++

	if s.nextReadIndex >= len(s.pending) {
		s.nextReadIndex = 0
	}

	return item
}

func (s *Subscription) processPending(proc goprocess.Process) {
	ctx := goprocessctx.OnClosingContext(proc)

	for {
		item := s.nextPending(true)

		if item == nil {
			return
		}

		func() {
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			if err := s.handler(ctx, *item); err != nil {
				panic(err)
			}
		}()
	}
}

func (s *Subscription) run(proc goprocess.Process) {
	proc.SetTeardown(s.teardown)
	proc.Go(s.processPending)

	defer func() {
		s.closed = true
		s.cond.Broadcast()
	}()

	for {
		select {
		case <-proc.Closing():
			return
		case msg := <-s.queue:
			s.appendPending(&msg)
		}
	}
}

func (s *Subscription) dispatch(msg Notification) {
	if s.closed {
		return
	}

	s.queue <- msg
}

func (s *Subscription) IsCompatibleWith(msg Notification) bool {
	actual, err := msg.Path.RelativeTo(s.path)

	if err != nil {
		return false
	}

	if actual.Len() > s.depth {
		return false
	}

	return true
}

func (s *Subscription) Close() error {
	return s.proc.Close()
}

func (s *Subscription) teardown() error {
	if s.topic != nil {
		s.topic.removeSubscription(s)
		s.topic = nil
	}

	return nil
}
