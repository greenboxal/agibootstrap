package vfs

import (
	"context"
	"io/fs"
	"sync"
)

type Manager struct {
	fs FileSystem

	mu    sync.RWMutex
	fsMap map[string]*fileSystem
}

func NewManager() (*Manager, error) {
	return &Manager{
		fsMap: map[string]*fileSystem{},
	}, nil
}

func (m *Manager) CreateLocalFS(path string, options ...FileSystemOption) (FileSystem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if fsys, ok := m.fsMap[path]; ok {
		return fsys, nil
	}

	fsys, err := newLocalFS(m, path, options...)

	if err != nil {
		return nil, err
	}

	m.fsMap[path] = fsys

	return fsys, nil
}

func (m *Manager) getFsForPath(p string) (n *fileSystem, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for fsRoot, fsys := range m.fsMap {
		ok, err := isChildPath(fsRoot, p)

		if err != nil {
			return nil, err
		}

		if ok {
			return fsys, nil
		}
	}

	return nil, fs.ErrNotExist
}

func (m *Manager) GetNodeForPath(ctx context.Context, p string) (n Node, err error) {
	fsys, err := m.getFsForPath(p)

	if err != nil {
		return nil, err
	}

	return fsys.GetNodeForPath(ctx, p)
}

func (m *Manager) Stop(ctx context.Context) error {
	for _, fsys := range m.fsMap {
		if err := fsys.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) notifyClose(fs FileSystem) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range m.fsMap {
		if v == fs {
			delete(m.fsMap, k)
		}
	}
}
