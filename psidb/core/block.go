package core

import (
	"context"
	"sync"

	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/adapters/psidsadapter"
	graphfs "github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

func NewBlockManager(
	lc fx.Lifecycle,
	cfg *coreapi.Config,
	ds coreapi.DataStore,
) *BlockManager {
	bm := &BlockManager{
		blocks: map[string]graphfs.SuperBlock{},
		cfg:    cfg,
		ds:     ds,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return bm.Initialize(ctx)
		},

		OnStop: func(ctx context.Context) error {
			return bm.Close()
		},
	})

	return bm
}

type BlockManager struct {
	mu     sync.RWMutex
	blocks map[string]graphfs.SuperBlock

	cfg *coreapi.Config
	ds  coreapi.DataStore
}

func (bm *BlockManager) Initialize(ctx context.Context) error {
	sb, err := psidsadapter.NewDataStoreSuperBlock(ctx, bm.ds, bm.cfg.RootUUID, false)

	if err != nil {
		return err
	}

	bm.Register(sb)

	return nil
}

func (bm *BlockManager) Resolve(ctx context.Context, uuid string) (graphfs.SuperBlock, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if sb := bm.blocks[uuid]; sb != nil {
		return sb, nil
	}

	return nil, nil
}

func (bm *BlockManager) Register(sb graphfs.SuperBlock) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.blocks[sb.UUID()] = sb
}

func (bm *BlockManager) Close() error {
	return nil
}
