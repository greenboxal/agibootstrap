package session

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type Manager struct {
	mu sync.RWMutex

	srm  *inject.ServiceRegistrationManager
	core coreapi.Core

	sessions map[string]*Session
}

func NewManager(
	lc fx.Lifecycle,
	core coreapi.Core,
	srm *inject.ServiceRegistrationManager,
) coreapi.SessionManager {
	sm := &Manager{
		srm:      srm,
		core:     core,
		sessions: map[string]*Session{},
	}

	lc.Append(fx.Hook{
		OnStop: sm.Stop,
	})

	return sm
}

func (sm *Manager) CreateSession() coreapi.Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return sm.createSessionUnlocked(coreapi.SessionConfig{SessionID: uuid.NewString()})
}

func (sm *Manager) GetSession(id string) coreapi.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.sessions[id]
}

func (sm *Manager) GetOrCreateSession(cfg coreapi.SessionConfig) coreapi.Session {
	id := cfg.SessionID

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if s := sm.sessions[id]; s != nil {
		return s
	}

	if id == "" && len(sm.sessions) > 0 {
		return nil
	}

	return sm.createSessionUnlocked(cfg)
}

func (sm *Manager) createSessionUnlocked(cfg coreapi.SessionConfig) coreapi.Session {
	var sess, parent *Session

	if cfg.ParentSessionID != cfg.SessionID {
		parent = sm.sessions[cfg.ParentSessionID]
	}

	if parent != nil {
		cfg.Journal = nil
		cfg.Checkpoint = nil
		cfg.MetadataStore = nil
		cfg.GraphStore = nil

		sess = parent
	} else {
		sess = NewSession(sm, nil, cfg)
	}

	sm.sessions[sess.UUID()] = sess

	return sess
}

func (sm *Manager) onSessionStarted(sess *Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sessions[sess.UUID()] = sess

	parentId := ""

	if sess.parent != nil {
		parentId = sess.parent.UUID()
	}

	sess.logger.Infow("Session started", "session_id", sess.UUID(), "parent_session_id", parentId)
}

func (sm *Manager) onSessionFinish(sess *Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, sess.UUID())
	parentId := ""

	if sess.parent != nil {
		parentId = sess.parent.UUID()
	}

	sess.logger.Infow("Session closed", "session_id", sess.UUID(), "parent_session_id", parentId)
}

func (sm *Manager) Stop(ctx context.Context) error {
	if len(sm.sessions) > 0 {
		logger.Infow("Terminating open sessions")

		for len(sm.sessions) > 0 {
			var wg sync.WaitGroup

			sm.mu.Lock()

			for _, sess := range sm.sessions {
				sess := sess

				wg.Add(1)

				go func() {
					defer wg.Done()

					if err := sess.ShutdownAndWait(ctx); err != nil {
						logger.Error(err)
					}
				}()
			}

			sm.mu.Unlock()
			wg.Wait()
			sm.mu.Lock()
		}

		sm.mu.Unlock()
	}

	return nil
}
