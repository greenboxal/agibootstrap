package session

import (
	"sync"

	"github.com/google/uuid"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type Manager struct {
	mu       sync.RWMutex
	core     coreapi.Core
	sessions map[string]*Session
}

func NewManager(core coreapi.Core) coreapi.SessionManager {
	return &Manager{
		core:     core,
		sessions: map[string]*Session{},
	}
}

func (m *Manager) CreateSession() coreapi.Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.createSessionUnlocked(uuid.NewString())
}

func (m *Manager) GetSession(id string) coreapi.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.sessions[id]
}

func (m *Manager) onSessionFinish(s *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, s.UUID())
}

func (m *Manager) GetOrCreateSession(id string) coreapi.Session {
	if id == "" {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if s := m.sessions[id]; s != nil {
		return s
	}

	return m.createSessionUnlocked(id)
}

func (m *Manager) createSessionUnlocked(id string) coreapi.Session {
	sess := NewSession(m, id)

	m.sessions[sess.UUID()] = sess

	return sess
}
