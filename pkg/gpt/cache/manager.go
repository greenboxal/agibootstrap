package cache

import (
	"os"

	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger"
)

type EmbeddingCacheManager struct {
	ds datastore.Batching
}

func NewEmbeddingCacheManager(path string) (*EmbeddingCacheManager, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	opts := badger.DefaultOptions
	ds, err := badger.NewDatastore(path, &opts)

	if err != nil {
		return nil, err
	}

	return &EmbeddingCacheManager{
		ds: ds,
	}, nil
}

func (m *EmbeddingCacheManager) GetEmbedder(embedder llm.Embedder) llm.Embedder {
	return NewCachedEmbedder(m.ds, embedder)
}

func (m *EmbeddingCacheManager) Close() error {
	return m.ds.Close()
}
