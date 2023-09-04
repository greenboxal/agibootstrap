package session

import (
	"sync"

	"github.com/google/uuid"

	`github.com/greenboxal/agibootstrap/pkg/platform/inject`
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type Manager struct {
	mu sync.RWMutex

	srm  *inject.ServiceRegistrationManager
	core coreapi.Core

	sessions map[string]*Session
}

func NewManager(core coreapi.Core, srm *inject.ServiceRegistrationManager) coreapi.SessionManager {
	return &Manager{
		srm:      srm,
		core:     core,
		sessions: map[string]*Session{},
	}
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
		cfg.LinkedStore = nil
		cfg.MetadataStore = nil

		sess = parent.Fork(cfg).(*Session)
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
