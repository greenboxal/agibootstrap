package fti

import (
	"context"
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
	for _, idx := range srcs {
		// Query the index and get the hits
		hits, err := idx.QueryClosestHits(r.context, r.query, r.k)
		if err != nil {
			return err
		}

		// Rerank the hits into the temporary FAISS index
		for i, hit := range hits {
			// Convert the hit to a FAISS embedding
			embeddings := hit.Entry.Embeddings.Float32()
			emb := llm.Embedding{Embeddings: embeddings}

			// Add the embedding to the temporary index
			if err := r.temp.Add(emb); err != nil {
				return err
			}

			// Update the hit with the new index and distance
			hits[i].Entry.Index = i
			hits[i].Distance = r.temp.CalculateDistance(emb)
		}

		// Sort the hits by distance
		faiss.SortByDistance(hits)

		// Replace the original index with the reranked hits
		idx.ReplaceEntries(hits)
	}

	return nil

}

// TODO: Implement this by querying each index in srcs and then reranking the results into FAISS index r.temp .
func (r *Index) RerankResults(srcs []*Index) error {
	// Query each index in srcs and get the results
	results := make([][]indexing.SearchHit, len(srcs))
	for i, src := range srcs {
		hits, err := src.QueryClosestHits()
		if err != nil {
			return err
		}
		results[i] = hits
	}

	// Rerank the results into the FAISS index r.temp
	r.temp = make([]indexing.SearchHit, 0)
	for _, hits := range results {
		r.temp = append(r.temp, hits...)
	}

	// Sort the reranked results by distance
	sort.Slice(r.temp, func(i, j int) bool {
		return r.temp[i].Distance < r.temp[j].Distance
	})

	return nil
}
func (r *Repository) ReRankResults(srcs []*Repository) error {
	for _, src := range srcs {
		hits, err := src.Query(r.tempQuery, DefaultK)
		if err != nil {
			return err
		}

		for _, hit := range hits {
			if err := r.temp.Add(hit); err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO: Implement this by querying each index in srcs and then reranking the results into FAISS index r.temp .
func (r *Index) RerankIndexes(ctx context.Context, srcs []*Index) error {
	// Create a temporary FAISS index
	tempIndex, err := faiss.NewFlatL2(r.dim)
	if err != nil {
		return err
	}

	// Loop through each source index
	for _, src := range srcs {
		// Query the source index
		hits, err := src.Query(ctx, r.query, r.k)
		if err != nil {
			return err
		}
		// Rerank the results and add them to the temporary index
		for _, hit := range hits {
			// TODO: Implement reranking logic here
			// Add the reranked result to the temporary index
			err = tempIndex.Add(hit.Embedding)
			if err != nil {
				return err
			}
		}
	}

	// Replace the current index with the temporary index
	r.index = tempIndex

	return nil
}
