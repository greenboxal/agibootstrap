package indexing

import (
	"context"
	"os"
	"path"
	"sync"

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

	openIndexes map[string]*referenceCountingIndex

	stream *coreapi.ReplicationStreamProcessor
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

		openIndexes: make(map[string]*referenceCountingIndex),
	}

	lc.Append(fx.Hook{
		OnStart: im.Start,
		OnStop:  im.Close,
	})

	//core.VirtualGraph().SetListener(im)

	return im, nil
}

func (im *Manager) Start(ctx context.Context) error {
	slot, err := im.core.CreateReplicationSlot(ctx, "indexmanager")

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

		dirtyNodes[entry.Path.String()] = *entry.Path
	}

	if len(dirtyNodes) == 0 {
		return nil
	}

	keys := maps.Values(dirtyNodes)

	slices.SortFunc(keys, func(i, j psi.Path) bool {
		return i.CompareTo(j) > 0
	})

	return im.Update(ctx, keys)
}

func (im *Manager) OpenNodeIndex(ctx context.Context, id string, embedder NodeEmbedder) (NodeIndex, error) {
	basic, err := im.OpenBasicIndex(ctx, id, embedder.Dimensions())

	if err != nil {
		return nil, err
	}

	return &nodeIndex{
		manager:  im,
		embedder: embedder,
		index:    basic,
	}, nil
}

func (im *Manager) OpenBasicIndex(ctx context.Context, id string, d int) (result BasicIndex, err error) {
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
		return idx, nil
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
	return im.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
		for _, p := range paths {
			if err := im.updateSingle(ctx, tx, p); err != nil {
				im.logger.Error(err)
			}
		}

		return nil
	})
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
	im.mu.Lock()
	defer im.mu.Unlock()

	for _, idx := range im.openIndexes {
		if err := idx.Close(); err != nil {
			return err
		}
	}

	return im.stream.Close(ctx)
}

func (im *Manager) notifyIndexIdle(id string) {
}
