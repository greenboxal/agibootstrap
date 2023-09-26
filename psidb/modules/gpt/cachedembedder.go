package gpt

import (
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/ipfs/go-datastore"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/psids"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type CachedEmbedder struct {
	llm.Embedder
	ds datastore.Batching
}

func NewCachedEmbedder(ds datastore.Batching, embedder llm.Embedder) *CachedEmbedder {
	return &CachedEmbedder{
		Embedder: embedder,
		ds:       ds,
	}
}

func (c *CachedEmbedder) GetEmbeddings(ctx context.Context, chunks []string) ([]llm.Embedding, error) {
	var uncachedIndexes []int

	embeddings := make([]llm.Embedding, len(chunks))
	hashes := make([]multihash.Multihash, len(chunks))

	for i, chunk := range chunks {
		if chunk == "" {
			embeddings[i] = llm.Embedding{Embeddings: make([]float32, c.Dimensions())}
			continue
		}

		mh, err := multihash.Sum([]byte(chunk), multihash.SHA2_256, -1)

		if err != nil {
			return nil, err
		}

		hashes[i] = mh

		cached, err := c.lookup(ctx, mh)

		if err != nil {
			if errors.Is(err, psi.ErrNodeNotFound) {
				uncachedIndexes = append(uncachedIndexes, i)
			} else {
				return nil, err
			}
		} else {
			embeddings[i] = cached
		}
	}

	if len(uncachedIndexes) > 0 {
		uncachedChunks := make([]string, len(uncachedIndexes))

		for i, j := range uncachedIndexes {
			uncachedChunks[i] = chunks[j]
		}

		uncachedEmbeddings, err := c.Embedder.GetEmbeddings(ctx, uncachedChunks)

		if err != nil {
			return nil, err
		}

		for i, j := range uncachedIndexes {
			embeddings[j] = uncachedEmbeddings[i]

			if err := c.put(ctx, hashes[j], embeddings[j]); err != nil {
				return nil, err
			}
		}
	}

	return embeddings, nil
}

func (c *CachedEmbedder) lookup(ctx context.Context, mh multihash.Multihash) (llm.Embedding, error) {
	return psids.Get(ctx, c.ds, dsKeyEmbedding(c.Embedder.Identity(), mh.B58String()))
}

func (c *CachedEmbedder) put(ctx context.Context, mh multihash.Multihash, embedding llm.Embedding) error {
	return psids.Put(ctx, c.ds, dsKeyEmbedding(c.Embedder.Identity(), mh.B58String()), embedding)
}

var dsKeyEmbedding = psids.KeyTemplate[llm.Embedding]("_embeddings/%s/%s")
