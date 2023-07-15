package vfs

import (
	"context"
	"io/fs"
	"os"
	"sync"

	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger"
)

type Manager struct {
	fs FileSystem
	ds datastore.Batching

	mu    sync.RWMutex
	fsMap map[string]*fileSystem
}

func NewManager(cachePath string) (*Manager, error) {
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return nil, err
	}

	opts := badger.DefaultOptions
	ds, err := badger.NewDatastore(cachePath, &opts)

	if err != nil {
		return nil, err
	}

	return &Manager{
		ds:    ds,
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

func (m *Manager) GetNodeForPath(ctx context.Context, p string) (n Node, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for fsRoot, fsys := range m.fsMap {
		ok, err := isChildPath(fsRoot, p)

		if err != nil {
			return nil, err
		}

		if ok {
			return fsys.GetNodeForPath(ctx, p)
		}
	}

	return nil, fs.ErrNotExist
}

func (m *Manager) Shutdown(ctx context.Context) error {
	for _, fsys := range m.fsMap {
		if err := fsys.Close(); err != nil {
			return err
		}
	}

	return m.ds.Close()
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
