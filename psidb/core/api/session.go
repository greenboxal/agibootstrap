package coreapi

import (
	"context"
	"time"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
)

type SessionKey string

type Session interface {
	SessionID() uint64
	SessionKey() SessionKey

	ServiceProvider() inject.ServiceProvider

	TransactionOperations

	Close() error
}

type SessionOption func(*SessionOptions)

type SessionOptions struct {
	Key SessionKey

	Deadline  *time.Time
	Timeout   *time.Duration
	KeepAlive *time.Duration
}

func (s *SessionOptions) Apply(options ...SessionOption) {
	for _, option := range options {
		option(s)
	}
}

func WithSessionDeadline(deadline time.Time) SessionOption {
	return func(options *SessionOptions) {
		options.Deadline = &deadline
	}
}

func WithSessionTimeout(timeout time.Duration) SessionOption {
	return func(options *SessionOptions) {
		options.Timeout = &timeout
	}
}

func WithSessionKeepAlive(keepAlive time.Duration) SessionOption {
	return func(options *SessionOptions) {
		options.KeepAlive = &keepAlive
	}
}

func WithSessionKey(key SessionKey) SessionOption {
	return func(options *SessionOptions) {
		options.Key = key
	}
}

type SessionManager interface {
	NewSession(ctx context.Context, options ...SessionOption) (Session, error)
	GetSession(ctx context.Context, key SessionKey) (Session, error)
}
