package fti

import (
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"

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
	// TODO: Implement this by querying each index in srcs and then reranking the results into r.temp.
	// Query each index in srcs
	for _, src := range r.srcs {
		hits, err := src.Query(q, k)
		if err != nil {
			return nil, err
		}
		// Implement the reranking logic here
	}
	return nil, nil
}

// Implement this by querying each index in srcs and then reranking the results into r.temp.
func rerankResults(srcs []index, query string, k int) {
	r.temp = make([]Result, 0)
	for _, src := range srcs {
		hits, err := src.Query(query, k)
		if err != nil {
			// handle error
		}
		r.temp = append(r.temp, hits...)
	}

	// Rerank results based on some criteria
	// ...

	// Sort the results
	// ...

	// Return the reranked results
}
func orphanSnippet() {
	// TODO: Implement this by querying each index in srcs and then reranking the results into r.temp.

	for _, index := range srcs {
		// Query each index in srcs
		hits, err := index.Query(ctx, query, k)
		if err != nil {
			return err
		}

		// Rerank the results into r.temp
		r.temp = append(r.temp, hits...)
	}

	// Rerank the results
	// ...

	return nil
}
