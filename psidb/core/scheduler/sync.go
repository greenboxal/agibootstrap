package scheduler

import (
	"sync"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type SemaphoreKind int

const (
	SemaphoreKindBinary   SemaphoreKind = iota
	SemaphoreKindCounting SemaphoreKind = iota
)

type SyncManager struct {
	mu         sync.RWMutex
	semaphores map[psi.PromiseHandle]coreapi.Semaphore

	scheduler *Scheduler
}

func NewSyncManager(
	sch *Scheduler,
) *SyncManager {
	return &SyncManager{
		scheduler:  sch,
		semaphores: map[psi.PromiseHandle]coreapi.Semaphore{},
	}
}

func (m *SyncManager) GetOrCreateSemaphore(handle psi.PromiseHandle) coreapi.Semaphore {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sem, ok := m.semaphores[handle]; ok {
		return sem
	}

	sem := NewCountingSemaphore(m.scheduler, 0)

	m.semaphores[handle] = sem

	return sem
}

func (m *SyncManager) DeleteSemaphore(handle psi.PromiseHandle) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.semaphores, handle)
}
