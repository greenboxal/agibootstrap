package pubsub

import (
	"github.com/jbenet/goprocess"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type SubscriptionPattern struct {
	ID    string   `json:"id"`
	Path  psi.Path `json:"path"`
	Depth int      `json:"depth"`
}

type Subscription struct {
	topic   *Topic
	pattern SubscriptionPattern

	closed bool

	ch      chan Notification
	handler func(notification Notification)

	proc goprocess.Process
}

func (s *Subscription) ID() string                   { return s.pattern.ID }
func (s *Subscription) Pattern() SubscriptionPattern { return s.pattern }
func (s *Subscription) Topic() *Topic                { return s.topic }

func (s *Subscription) Push(n Notification) {
	if s.closed {
		return
	}

	s.ch <- n
}

func (s *Subscription) Close() error {
	if s.closed {
		return nil
	}

	close(s.ch)

	return s.proc.Close()
}

func (s *Subscription) pump(proc goprocess.Process) {
	defer func() {
		s.closed = true
		s.topic.Unsubscribe(s.pattern.ID)
	}()

	for {
		select {
		case <-proc.Closing():
			return
		case n, ok := <-s.ch:
			if !ok {
				return
			}

			s.handler(n)
		}
	}
}
