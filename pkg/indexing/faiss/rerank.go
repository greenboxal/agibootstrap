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
	// TODO: Write GoDoc for this function
	return &RerankIndex[K]{
		srcs: sources,
	}
}

// Query searches for files in the repository that are similar to the provided query.
// It takes a context, which can be used for cancellation, the query string, and the maximum number of results (k) to return.
// The function returns a slice of OnlineIndexQueryHit, which contains information about the matching files, and an error, if any.
func (r *RerankIndex[K]) Query(q llm.Embedding, k int64) ([]indexing.SearchHit[K], error) {
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
