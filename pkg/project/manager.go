package project

import "sync"

type Manager struct {
	mu       sync.RWMutex
	projects map[string]Project
}

func NewManager() *Manager {
	return &Manager{
		projects: map[string]Project{},
	}
}

func (m *Manager) AddProject(p Project) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.projects[p.UUID()] = p
}
