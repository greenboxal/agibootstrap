package jukebox

import (
	"sync"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	`github.com/greenboxal/agibootstrap/psidb/psi`
)

type QueueManager struct {
	mu sync.RWMutex

	cm   *ConnectionManager
	core coreapi.Core

	queues map[string]*PlayerQueue
}

func NewQueueManager(
	core coreapi.Core,
	cm *ConnectionManager,
) *QueueManager {
	return &QueueManager{
		cm:   cm,
		core: core,

		queues: map[string]*PlayerQueue{},
	}
}

func (m *QueueManager) GetOrCreateQueue(p psi.Path) *PlayerQueue {
	name := p.String()

	m.mu.RLock()
	if existing := m.queues[name]; existing != nil {
		m.mu.RUnlock()
		return existing
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if existing := m.queues[name]; existing != nil {
		return existing
	}

	player := NewPlayerQueue(m, p)

	m.queues[name] = player

	return player
}
