package gpt

import (
	"context"
	"os"
	"path"

	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger"
	"go.uber.org/fx"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type EmbeddingCacheManager struct {
	ds datastore.Batching
}

func NewEmbeddingCacheManager(
	lc fx.Lifecycle,
	cfg *coreapi.Config,
) (*EmbeddingCacheManager, error) {
	cacheDirPath := path.Join(cfg.DataDir, "embedding-cache")

	if err := os.MkdirAll(cacheDirPath, 0755); err != nil {
		return nil, err
	}

	opts := badger.DefaultOptions
	ds, err := badger.NewDatastore(cacheDirPath, &opts)

	if err != nil {
		return nil, err
	}

	ecm := &EmbeddingCacheManager{
		ds: ds,
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return ecm.Close()
		},
	})

	return ecm, nil
}

func (m *EmbeddingCacheManager) GetEmbedder(embedder llm.Embedder) llm.Embedder {
	return NewCachedEmbedder(m.ds, embedder)
}

func (m *EmbeddingCacheManager) Close() error {
	return m.ds.Close()
}
