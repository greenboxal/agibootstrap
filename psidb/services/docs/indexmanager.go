package docs

import (
	"context"
	"path"
	"sync"

	"go.uber.org/fx"
	"golang.org/x/exp/maps"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type IndexManager struct {
	workspacesPath string

	mu      sync.RWMutex
	indexes map[string]*LiveIndex
}

func NewIndexManager(
	lc fx.Lifecycle,
	cfg *coreapi.Config,
) *IndexManager {
	im := &IndexManager{
		indexes:        map[string]*LiveIndex{},
		workspacesPath: path.Join(cfg.DataDir, "indexes"),
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return im.Close()
		},
	})

	return im
}

func (im *IndexManager) GetOrCreateLiveIndex(uuid string) *LiveIndex {
	im.mu.RLock()
	idx := im.indexes[uuid]
	im.mu.RUnlock()

	if idx != nil {
		return idx
	}

	im.mu.Lock()
	defer im.mu.Unlock()

	idx = im.indexes[uuid]
	if idx != nil {
		return idx
	}

	idx, err := NewLiveIndex(im, uuid, path.Join(im.workspacesPath, uuid))

	if err != nil {
		panic(err)
	}

	im.indexes[uuid] = idx

	return idx
}

func (im *IndexManager) Close() error {
	for len(im.indexes) > 0 {
		im.mu.RLock()
		targets := maps.Values(im.indexes)
		im.mu.RUnlock()

		for _, idx := range targets {
			if err := idx.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (im *IndexManager) notifyClose(idx *LiveIndex) {
	im.mu.Lock()
	defer im.mu.Unlock()

	delete(im.indexes, idx.uuid)
}
