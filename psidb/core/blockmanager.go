package core

import (
	"context"
	"os"
	"path"
	"sync"

	badger "github.com/ipfs/go-ds-badger"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/psidb/core/api"
	graphfs2 "github.com/greenboxal/agibootstrap/psidb/db/graphfs"
	"github.com/greenboxal/agibootstrap/psidb/providers/psidsadapter"
)

func NewBlockManager(
	cfg *coreapi.Config,
	ds coreapi.DataStore,
) *BlockManager {
	bm := &BlockManager{
		blocks: map[string]graphfs2.SuperBlock{},
	}

	sb := psidsadapter.NewDataStoreSuperBlock(ds, cfg.RootUUID, false)

	bm.Register(sb)

	return bm
}

type BlockManager struct {
	mu     sync.RWMutex
	blocks map[string]graphfs2.SuperBlock
}

func (m *BlockManager) Resolve(ctx context.Context, uuid string) (graphfs2.SuperBlock, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if sb := m.blocks[uuid]; sb != nil {
		return sb, nil
	}

	return nil, nil
}

func (m *BlockManager) Register(sb graphfs2.SuperBlock) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blocks[sb.UUID()] = sb
}

func NewDataStore(
	lc fx.Lifecycle,
	cfg *coreapi.Config,
) (coreapi.DataStore, error) {
	dsOpts := badger.DefaultOptions
	dsPath := path.Join(cfg.DataDir, "data")

	if err := os.MkdirAll(dsPath, 0755); err != nil {
		return nil, err
	}

	ds, err := badger.NewDatastore(dsPath, &dsOpts)

	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return ds.Close()
		},
	})

	return ds, nil
}
