package fti

import (
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"

	"github.com/greenboxal/agibootstrap/pkg/indexing"
)

type RerankIndex[K comparable] struct {
	srcs []indexing.Index[K]
}

// TODO: Add package-level comment

// NewRerankIndex creates a new RerankIndex with the provided sources.
func NewRerankIndex[K comparable](sources ...indexing.Index[K]) *RerankIndex[K] {
	return &RerankIndex[K]{
		srcs: sources,
	}
}

// Query queries the RerankIndex with the provided embedding vector.
// TODO: Add detailed function description, parameters, and return values.
func (r *RerankIndex[K]) Query(q llm.Embedding, k int64) ([]indexing.SearchHit[K], error) {
	// TODO: Add implementation details and explanation of the algorithm used.
	// TODO: Add example usage if applicable.
	return nil, nil
}
