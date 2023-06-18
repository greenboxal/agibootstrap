package fti

import (
	"sort"

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
	// Query each index in srcs
	for _, index := range r.srcs {
		hits, err := index.Query(query, k)
		if err != nil {
			return nil, err
		}

		// Rerank the results into r.temp
		for _, hit := range hits {
			r.temp = append(r.temp, hit)
		}
	}

	// Sort the results in r.temp by score (distance)
	sort.Slice(r.temp, func(i, j int) bool {
		return r.temp[i].Distance < r.temp[j].Distance
	})

	// Trim the results to the desired maximum number of results (k)
	if int64(len(r.temp)) > k {
		r.temp = r.temp[:k]
	}

	return r.temp, nil

}
