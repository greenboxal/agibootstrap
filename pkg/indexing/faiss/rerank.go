package fti

import (
	"sync"

	"github.com/greenboxal/aip/aip-langchain/pkg/llm"

	"github.com/greenboxal/agibootstrap/pkg/indexing"
)

type RerankIndex[K comparable] struct {
	srcs []indexing.Index[K]
}

func NewRerankIndex[K comparable](sources ...indexing.Index[K]) *RerankIndex[K] {
	// TODO: Write documentation
	return &RerankIndex[K]{
		srcs: sources,
	}
}

func (r *RerankIndex[K]) Query(q llm.Embedding, k int64) ([]indexing.SearchHit[K], error) {
	// TODO: Write documentation

	// Add meaningful documentation for the Query function.
	// This function queries the RerankIndex with a given query embedding and returns a list of hits.
	// Each hit contains the corresponding entry from the index and the distance between the query and the entry embedding.

	var wg sync.WaitGroup

	temp, err := NewFlatKVIndex[K]()

	if err != nil {
		return nil, err
	}

	defer temp.Close()

	for _, src := range r.srcs {
		wg.Add(1)

		go func(src indexing.Index[K]) {
			defer wg.Done()

			hits, err := src.Query(q, k)

			if err != nil {
				return
			}

			for _, hit := range hits {
				_ = temp.Add(hit.DocumentID, hit.Embedding)
			}
		}(src)
	}

	// Return the query results
	return temp.Query(q, k)
}
func orphanSnippet() {
	// TODO: Implement query logic
	return nil, nil

}
