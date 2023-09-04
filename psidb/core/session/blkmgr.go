package session

import (
	`context`
	`sync`

	coreapi `github.com/greenboxal/agibootstrap/psidb/core/api`
	`github.com/greenboxal/agibootstrap/psidb/db/graphfs`
)

type BlockManager struct {
	mu sync.RWMutex

	mountTab     map[string]coreapi.MountDefinition
	activeMounts map[string]graphfs.SuperBlock
}

func NewBlockManager() *BlockManager {
	return &BlockManager{
		mountTab:     map[string]coreapi.MountDefinition{},
		activeMounts: map[string]graphfs.SuperBlock{},
	}
}

func (m *BlockManager) Resolve(ctx context.Context, uuid string) (graphfs.SuperBlock, error) {
	m.mu.RLock()

	if sb := m.activeMounts[uuid]; sb != nil {
		m.mu.RUnlock()
		return sb, nil
	}

	if md, ok := m.mountTab[uuid]; ok {
		m.mu.RUnlock()
		return m.Mount(ctx, md)
	}

	m.mu.RUnlock()

	return nil, nil
}

func (m *BlockManager) RegisterMountDefinition(md coreapi.MountDefinition) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mountTab[md.Name] = md
}

func (m *BlockManager) RegisterSuperBlock(name string, sb graphfs.SuperBlock) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.activeMounts[name] = sb
}

func (m *BlockManager) Mount(ctx context.Context, md coreapi.MountDefinition) (graphfs.SuperBlock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	uuid := md.Name

	if sb := m.activeMounts[uuid]; sb != nil {
		return sb, nil
	}

	sba, err := md.Target.Mount(ctx, md)

	if err != nil {
		return nil, err
	}

	sb := sba.(graphfs.SuperBlock)

	m.activeMounts[uuid] = sb

	return sb, nil
}

func (m *BlockManager) Unmount(ctx context.Context, sb graphfs.SuperBlock) error {
	uuid := sb.UUID()

	m.mu.RLock()
	existing := m.activeMounts[uuid]
	m.mu.RUnlock()

	if existing == nil {
		return nil
	} else if existing != sb {
		return coreapi.ErrUnsupportedOperation
	}

	if err := sb.Close(ctx); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if existing := m.activeMounts[uuid]; existing == sb {
		delete(m.activeMounts, uuid)
	}

	return nil
}

func (m *BlockManager) UnmountAll(ctx context.Context) error {
	for {
		var sb graphfs.SuperBlock

		m.mu.Lock()
		if len(m.activeMounts) == 0 {
			m.mu.Unlock()
			break
		}

		for _, sb = range m.activeMounts {
			break
		}

		m.mu.Unlock()

		if err := m.Unmount(ctx, sb); err != nil {
			return err
		}
	}

	return nil
}
