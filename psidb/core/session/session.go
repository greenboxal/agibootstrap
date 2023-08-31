package session

import (
	"context"
	"sync"
	"time"

	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
)

type Session struct {
	mu sync.RWMutex

	manager *Manager

	logger *otelzap.SugaredLogger

	uuid          string
	lastKeepAlive time.Time

	incomingMsgCh chan coreapi.SessionMessage
	outgoingMsgCh chan coreapi.SessionMessage

	stopCh chan struct{}
	proc   goprocess.Process

	clients []coreapi.SessionClient
}

func NewSession(m *Manager, id string) *Session {
	s := &Session{
		logger: logging.GetLogger("session"),

		manager: m,

		stopCh:        make(chan struct{}),
		incomingMsgCh: make(chan coreapi.SessionMessage, 16),
		outgoingMsgCh: make(chan coreapi.SessionMessage, 16),

		uuid: id,
	}

	s.proc = goprocess.Go(s.Run)

	return s
}

func (s *Session) UUID() string             { return s.uuid }
func (s *Session) KeepAlive()               { s.lastKeepAlive = time.Now() }
func (s *Session) LastKeepAlive() time.Time { return s.lastKeepAlive }

func (s *Session) ReceiveMessage(m coreapi.SessionMessage) {
	s.incomingMsgCh <- m
}

func (s *Session) SendMessage(m coreapi.SessionMessage) {
	s.outgoingMsgCh <- m
}

func (s *Session) AttachClient(client coreapi.SessionClient) {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := slices.Index(s.clients, client)

	if idx != -1 {
		return
	}

	s.clients = append(s.clients, client)
}

func (s *Session) DetachClient(client coreapi.SessionClient) {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := slices.Index(s.clients, client)

	if idx == -1 {
		return
	}

	s.clients = slices.Delete(s.clients, idx, idx+1)
}

func (s *Session) BeginTransaction(ctx context.Context, options ...coreapi.TransactionOption) (coreapi.Transaction, error) {
	ctx = coreapi.WithSession(ctx, s)
	return s.manager.core.BeginTransaction(ctx, options...)
}

func (s *Session) RunTransaction(ctx context.Context, fn coreapi.TransactionFunc, options ...coreapi.TransactionOption) (err error) {
	ctx = coreapi.WithSession(ctx, s)
	return coreapi.RunTransaction(ctx, s, fn, options...)
}

func (s *Session) Run(proc goprocess.Process) {
	defer s.manager.onSessionFinish(s)

	ctx := goprocessctx.OnClosingContext(proc)
	ticker := time.NewTicker(30 * time.Second)

	for {
		select {
		case _, _ = <-s.stopCh:
			return

		case <-ticker.C:
			if time.Now().Sub(s.lastKeepAlive) > 30*time.Second {
				close(s.stopCh)
			}

		case msg := <-s.outgoingMsgCh:
			if err := s.sendOutgoingMessage(ctx, msg); err != nil {
				s.logger.Error(err)
			}

		case msg := <-s.incomingMsgCh:
			if err := s.processMessage(ctx, msg); err != nil {
				s.logger.Error(err)
			}
		}
	}
}

func (s *Session) processMessage(ctx context.Context, msg coreapi.SessionMessage) error {
	s.lastKeepAlive = time.Now()

	switch msg.(type) {
	case coreapi.SessionMessageKeepAlive:
		return nil

	case coreapi.SessionMessageShutdown:
		return s.proc.Close()
	}

	return nil
}

func (s *Session) sendOutgoingMessage(ctx context.Context, msg coreapi.SessionMessage) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		if err := client.SendSessionMessage(s.UUID(), msg); err != nil {
			s.logger.Warn(err)
		}
	}

	return nil
}

func (s *Session) Close() error {
	close(s.stopCh)

	return s.proc.CloseAfterChildren()
}
