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
	return &RerankIndex[K]{
		srcs: sources,
	}
}

// Query performs a search in the RerankIndex using the provided query embedding and returns a list of hits.
// Each hit contains the corresponding entry from the index and the distance between the query and the entry embedding.
// The function takes the query embedding q and the number of desired hits k as input arguments.
// It returns a slice of SearchHit objects and an error, if any.

func (r *RerankIndex[K]) Query(q llm.Embedding, k int64) ([]indexing.SearchHit[K], error) {
	var wg sync.WaitGroup

	// Create a temporary FlatKVIndex to store the intermediate search results
	temp, err := NewFlatKVIndex[K]()

	if err != nil {
		return nil, err
	}

	defer temp.Close()

	// Perform a parallel query on each source index
	for _, src := range r.srcs {
		wg.Add(1)

		go func(src indexing.Index[K]) {
			defer wg.Done()

			// Perform the query on the source index
			hits, err := src.Query(q, k)

			if err != nil {
				return
			}

			// Add each hit to the temporary index
			for _, hit := range hits {
				_ = temp.Add(hit.DocumentID, hit.Embedding)
			}
		}(src)
	}

	// Wait for all queries to complete
	wg.Wait()

	// Perform the final query on the temporary index
	return temp.Query(q, k)
}
