package fti

import (
	"sort"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"

	"github.com/greenboxal/agibootstrap/pkg/indexing"
)

type RerankIndex[K comparable] struct {
	srcs []indexing.Index[K]
}

func NewRerankIndex[K comparable](sources ...indexing.Index[K]) *RerankIndex[K] {
	return &RerankIndex[K]{
		srcs: sources,
	}
}

func (r *RerankIndex[K]) Query(q llm.Embedding, k int64) ([]indexing.SearchHit[K], error) {
	temp, err := NewFlatKVIndex[K]()

	if err != nil {
		return nil, err
	}

	defer temp.Close()
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
func orphanSnippet() {
	// Query each source index and collect the results.
	var results []*faiss.Index

	for _, src := range srcs {
		hits, err := src.Query(ctx, r.temp, -1)

		if err != nil {
			return err
		}

		results = append(results, hits...)
	}

	// Rerank the results based on some criteria.

	// Sort the results based on distance.

	sort.Slice(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})

	// Add the reranked results to the target index.

	for _, hit := range results {
		if err := r.Add(hit.Embedding); err != nil {
			return err
		}
	}

	return nil

}
