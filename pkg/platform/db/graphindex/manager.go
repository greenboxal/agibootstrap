package graphindex

import (
	"context"
	"sync"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphstore"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
)

var logger = logging.GetLogger("graphindex")

type Manager struct {
	basePath string

	graph *graphstore.IndexedGraph

	mu          sync.RWMutex
	openIndexes map[string]*referenceCountingIndex
}

func NewManager(graph *graphstore.IndexedGraph, basePath string) *Manager {
	return &Manager{
		graph:       graph,
		basePath:    basePath,
		openIndexes: make(map[string]*referenceCountingIndex),
	}
}

func (m *Manager) OpenNodeIndex(ctx context.Context, id string, embedder NodeEmbedder) (NodeIndex, error) {
	basic, err := m.OpenBasicIndex(ctx, id, embedder.Dimensions())

	if err != nil {
		return nil, err
	}

	return &nodeIndex{
		graph:    m.graph,
		embedder: embedder,
		index:    basic,
	}, nil
}

func (m *Manager) OpenBasicIndex(ctx context.Context, id string, d int) (result BasicIndex, err error) {
	isNew := false

	defer func() {
		if err == nil && isNew {
			err = result.Load()

		}

		if err != nil {
			result = nil
		}
	}()

	m.mu.Lock()
	defer m.mu.Unlock()

	if idx, ok := m.openIndexes[id]; ok {
		return idx, nil
	}

	idx, err := newFaissIndex(m, m.basePath, id, d)

	if err != nil {
		return nil, err
	}

	refCounting := &referenceCountingIndex{faissIndex: idx}

	m.openIndexes[id] = refCounting

	isNew = true

	return refCounting.ref()
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, idx := range m.openIndexes {
		if err := idx.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) notifyIndexIdle(id string) {
}
