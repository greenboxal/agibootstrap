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
	// TODO: Implement this by querying each index in srcs and then reranking the results into temp.
	// Query each index in srcs
	for _, src := range r.srcs {
		hits, err := src.Query(q, k)
		if err != nil {
			return nil, err
		}
		// Rerank the results into temp
		temp := make([]indexing.SearchHit[K], len(hits))
		copy(temp, hits)
		// Implement the reranking logic here
		return temp, nil
	}
	return nil, nil
}
