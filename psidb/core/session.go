package core

import (
	"context"
	"io/fs"
	"sync"

	"github.com/jbenet/goprocess"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"golang.org/x/exp/maps"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type SessionManager struct {
	core *Core

	mu           sync.RWMutex
	sessionById  map[uint64]coreapi.Session
	sessionByKey map[coreapi.SessionKey]coreapi.Session

	sidCounter uint64

	closing bool
	proc    goprocess.Process
}

var SessionNotFound = errors.Wrap(psi.ErrNodeNotFound, "session not found")

func NewSessionManager(
	lc fx.Lifecycle,
	core *Core,
) coreapi.SessionManager {
	sm := &SessionManager{
		core: core,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return sm.Start(ctx)
		},

		OnStop: func(ctx context.Context) error {
			return sm.Shutdown(ctx)
		},
	})

	return sm
}

func (sm *SessionManager) NewSession(ctx context.Context, options ...coreapi.SessionOption) (coreapi.Session, error) {
	var opts coreapi.SessionOptions

	opts.Apply(options...)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if existing := sm.sessionByKey[opts.Key]; existing != nil {
		return nil, errors.Wrapf(fs.ErrExist, "session with key %s already exists", opts.Key)
	}

	sid := sm.sidCounter
	sm.sidCounter++

	sess := &Session{
		sid:  sid,
		skey: opts.Key,
		sp:   inject.NewServiceProvider(),
	}

	sm.sessionById[sid] = sess
	sm.sessionByKey[opts.Key] = sess

	if err := sess.Start(ctx); err != nil {
		return nil, err
	}

	return sess, nil
}

func (sm *SessionManager) GetSession(ctx context.Context, key coreapi.SessionKey) (coreapi.Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sess := sm.sessionByKey[key]

	if sess == nil {
		return nil, SessionNotFound
	}

	return sess, nil
}

func (sm *SessionManager) Start(ctx context.Context) error {
	return nil
}

func (sm *SessionManager) Shutdown(ctx context.Context) error {
	sm.mu.Lock()
	sm.closing = true
	sessions := maps.Values(sm.sessionById)
	sm.mu.Unlock()

	for _, sess := range sessions {
		if err := sess.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (sm *SessionManager) notifySessionClosed(s *Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessionById, s.sid)
	delete(sm.sessionByKey, s.skey)
}

type Session struct {
	mu   sync.RWMutex
	sid  uint64
	skey coreapi.SessionKey

	sp   inject.ServiceProvider
	sm   *SessionManager
	core *Core

	closed bool
	proc   goprocess.Process
}

func (s *Session) SessionID() uint64                       { return s.sid }
func (s *Session) SessionKey() coreapi.SessionKey          { return s.skey }
func (s *Session) ServiceProvider() inject.ServiceProvider { return s.sp }

func (s *Session) BeginTransaction(ctx context.Context, options ...coreapi.TransactionOption) (coreapi.Transaction, error) {
	options = append(options, coreapi.WithServiceLocator(s.sp))

	return s.core.BeginTransaction(ctx, options...)
}

func (s *Session) RunTransaction(ctx context.Context, fn coreapi.TransactionFunc, options ...coreapi.TransactionOption) (err error) {
	return coreapi.RunTransaction(ctx, s, fn, options...)
}

func (s *Session) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.proc != nil {
		return nil
	}

	s.proc = goprocess.Go(func(proc goprocess.Process) {
		defer s.onAfterClose()

		<-proc.Closing()
	})

	return nil
}

func (s *Session) Close() error {
	if s.proc != nil {
		if err := s.proc.CloseAfterChildren(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Session) onAfterClose() {
	// TODO: Shutdown service provider

	s.closed = true

	s.sm.notifySessionClosed(s)
}
