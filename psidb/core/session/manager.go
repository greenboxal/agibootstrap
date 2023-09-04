package session

import (
	`context`
	"sync"

	"github.com/google/uuid"

	`github.com/greenboxal/agibootstrap/pkg/platform/inject`
	`github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators`
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

func (m *Manager) CreateSession() coreapi.Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.createSessionUnlocked(coreapi.SessionConfig{SessionID: uuid.NewString()})
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

func (m *Manager) GetOrCreateSession(cfg coreapi.SessionConfig) coreapi.Session {
	id := cfg.SessionID

	m.mu.Lock()
	defer m.mu.Unlock()

	if s := m.sessions[id]; s != nil {
		return s
	}

	if id == "" && len(m.sessions) > 0 {
		return nil
	}

	return m.createSessionUnlocked(cfg)
}

func (m *Manager) createSessionUnlocked(cfg coreapi.SessionConfig) coreapi.Session {
	var sess, parent *Session

	if cfg.ParentSessionID != cfg.SessionID {
		parent = m.sessions[cfg.ParentSessionID]
	}

	if parent != nil {
		cfg.Journal = nil
		cfg.Checkpoint = nil
		cfg.LinkedStore = nil
		cfg.MetadataStore = nil

		sess = parent.Fork(cfg).(*Session)
	} else {
		sess = NewSession(m, nil, cfg)
	}

	m.sessions[sess.UUID()] = sess

	return sess
}

type sessionFramer struct {
	mu      sync.RWMutex
	head    uint64
	records []coreapi.JournalEntry
}

func (s *sessionFramer) GetHead() (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.head, nil
}

func (s *sessionFramer) Iterate(startIndex uint64, count int) iterators.Iterator[coreapi.JournalEntry] {
	return iterators.FromSlice(s.records[startIndex-1 : startIndex+uint64(count)-1])
}

func (s *sessionFramer) Read(index uint64, dst *coreapi.JournalEntry) (*coreapi.JournalEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	*dst = s.records[index-1]

	return dst, nil
}

func (s *sessionFramer) Write(op *coreapi.JournalEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.head++

	op.Rid = s.head
	s.records = append(s.records, *op)

	return nil
}

func (s *sessionFramer) Close() error {
	return nil
}

func (s *sessionFramer) CreateJournal(ctx context.Context) (coreapi.Journal, error) {
	return s, nil
}
