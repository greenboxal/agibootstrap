package pubsub

import "github.com/jbenet/goprocess"

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
	s.ch <- n
}

func (s *Subscription) Close() error {
	if s.closed {
		return nil
	}

	s.topic.Unsubscribe(s.pattern.ID)
	s.closed = true

	return nil
}

func (s *Subscription) pump(proc goprocess.Process) {
	defer close(s.ch)

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
