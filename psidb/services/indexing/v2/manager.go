package indexing

import (
	"context"
	"os"
	"path"
	"sync"

	"github.com/alitto/pond"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

type Manager struct {
	mu     sync.RWMutex
	logger *zap.SugaredLogger

	core     coreapi.Core
	basePath string

	openIndexes map[string]*LiveIndex

	stream *coreapi.ReplicationStreamProcessor

	pool *pond.WorkerPool
}

func NewIndexManager(
	lc fx.Lifecycle,
	core coreapi.Core,
) (*Manager, error) {
	indexPath := path.Join(core.Config().DataDir, "index")

	if err := os.MkdirAll(indexPath, 0755); err != nil {
		panic(err)
	}

	im := &Manager{
		logger: logging.GetLogger("indexmanager"),

		core:     core,
		basePath: indexPath,

		openIndexes: make(map[string]*LiveIndex),

		pool: pond.New(8, 1024),
	}

	lc.Append(fx.Hook{
		OnStart: im.Start,
		OnStop:  im.Close,
	})

	return im, nil
}

func (im *Manager) Start(ctx context.Context) error {
	slot, err := im.core.CreateReplicationSlot(ctx, graphfs.ReplicationSlotOptions{
		Name: "indexmanager",
	})

	if err != nil {
		return err
	}

	im.stream = coreapi.NewReplicationStream(slot, im.processReplicationMessage)

	return nil
}

func (im *Manager) processReplicationMessage(ctx context.Context, entries []*graphfs.JournalEntry) error {
	dirtyNodes := map[string]psi.Path{}

	for _, entry := range entries {
		if entry.Path == nil {
			continue
		}

		switch entry.Op {
		case graphfs.JournalOpWrite:
			fallthrough
		case graphfs.JournalOpSetEdge:
			dirtyNodes[entry.Path.String()] = *entry.Path

		case graphfs.JournalOpRemoveEdge:
			delete(dirtyNodes, entry.Path.String())
		}

	}

	if len(dirtyNodes) == 0 {
		return nil
	}

	keys := maps.Values(dirtyNodes)

	slices.SortFunc(keys, func(i, j psi.Path) bool {
		return i.CompareTo(j) < 0
	})

	return im.Update(ctx, keys)
}

func (im *Manager) OpenIndex(ctx context.Context, id string, d int) (result *LiveIndex, err error) {
	isNew := false

	defer func() {
		if err == nil && isNew {
			err = result.Load()
		}

		if err != nil {
			result = nil
		}
	}()

	im.mu.Lock()
	defer im.mu.Unlock()

	if idx, ok := im.openIndexes[id]; ok {
		return idx.ref()
	}

	idx, err := newFaissIndex(im, path.Join(im.basePath, id), id, d)

	if err != nil {
		return nil, err
	}

	refCounting := &referenceCountingIndex{faissIndex: idx}

	im.openIndexes[id] = refCounting

	isNew = true

	return refCounting.ref()
}

func (im *Manager) Update(ctx context.Context, paths []psi.Path) error {
	tg, gctx := im.pool.GroupContext(ctx)

	for _, p := range paths {
		p := p

		tg.Submit(func() error {
			return im.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
				if err := im.updateSingle(gctx, tx, p); err != nil {
					im.logger.Error(err)
				}

				return nil
			}, coreapi.WithReadOnly())
		})
	}

	return tg.Wait()
}

func (im *Manager) updateSingle(ctx context.Context, tx coreapi.Transaction, p psi.Path) error {
	im.logger.Infow("Updating index", "path", p)

	node, err := tx.Resolve(ctx, p)

	if err != nil {
		return err
	}

	scp := GetHierarchyScope(ctx, node)

	if node == scp || scp == nil {
		return nil
	}

	if err := scp.Upsert(ctx, node); err != nil {
		return err
	}

	return nil
}

func (im *Manager) Close(ctx context.Context) error {
	im.pool.StopAndWait()

	im.mu.Lock()
	defer im.mu.Unlock()

	for _, idx := range im.openIndexes {
		if err := idx.Close(true); err != nil {
			return err
		}
	}

	if err := im.stream.Close(ctx); err != nil {
		return err
	}

	return nil
}

func (im *Manager) notifyIndexIdle(ir *faissIndex) {
	if err := ir.Save(); err != nil {
		im.logger.Error(err)
	}
}
