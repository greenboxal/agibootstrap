package fti

import (
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/greenboxal/aip/aip-langchain/pkg/vectorstore/faiss"

	"github.com/greenboxal/agibootstrap/pkg/indexing"
)

type RerankIndex[K comparable] struct {
	srcs []indexing.Index[K]
	temp *FlatKVIndex[K]
}

func NewRerankIndex[K comparable](sources ...indexing.Index[K]) *RerankIndex[K] {
	return &RerankIndex[K]{
		srcs: []indexing.Index[K]{},
	}
}

func (r *RerankIndex[K]) Query(q llm.Embedding, k int64) ([]indexing.SearchHit[K], error) {
	// TODO: Implement this by querying each index in srcs and then reranking the results into FAISS index r.temp .
	// Query each index in srcs
	for _, src := range r.srcs {
		hits, err := src.Query(q, k)
		if err != nil {
			return nil, err
		}
		// Implement the reranking logic here
	}

	return r.temp.Query(q, k)
}
func (r *Repository) RerankResults(ctx context.Context, srcs []Index) error {
	// Create a temporary FAISS index for storing the reranked results
	temp, err := faiss.NewIndex(r.faissConfig)
	if err != nil {
		return err
	}

	// Query each index in srcs and add the results to the temporary index
	for _, src := range srcs {
		hits, err := src.Query(ctx, r.query, r.k)
		if err != nil {
			return err
		}

		// Rerank the hits and add them to the temporary index
		rerankedHits, err := r.rerankHits(hits)
		if err != nil {
			return err
		}

		for _, hit := range rerankedHits {
			if err := temp.Add(hit); err != nil {
				return err
			}
		}
	}

	// Set the r.temp index to the reranked results
	r.temp = temp

	return nil
}

func (r *Repository) rerankHits(hits []IndexHit) ([]IndexHit, error) {
	// Implement your reranking logic here

	// TODO: Implement the reranking logic to reorder the hits based on some criteria

	return hits, nil
}
