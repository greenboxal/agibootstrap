package faiss

import "C"
import (
	"context"
	"fmt"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/greenboxal/aip/aip-langchain/pkg/chunkers"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
)

//// faiss.Index is a Faiss index.
////
//// Note that some index implementations do not support all methods.
//// Check the Faiss wiki to see what operations an index supports.
//type faiss.Index interface {
//	// D returns the dimension of the indexed vectors.
//	D() int
//
//	// IsTrained returns true if the index has been trained or does not require
//	// training.
//	IsTrained() bool
//
//	// Ntotal returns the number of indexed vectors.
//	Ntotal() int64
//
//	// MetricType returns the metric type of the index.
//	MetricType() int
//
//	// Train trains the index on a representative set of vectors.
//	Train(x []float32) error
//
//	// Add adds vectors to the index.
//	Add(x []float32) error
//
//	// AddWithIDs is like Add, but stores xids instead of sequential IDs.
//	AddWithIDs(x []float32, xids []int64) error
//
//	// Search queries the index with the vectors in x.
//	// Returns the IDs of the k nearest neighbors for each query vector and the
//	// corresponding distances.
//	Search(x []float32, k int64) (distances []float32, labels []int64, err error)
//
//	// RangeSearch queries the index with the vectors in x.
//	// Returns all vectors with distance < radius.
//	RangeSearch(x []float32, radius float32) (*faiss.RangeSearchResult, error)
//
//	// Reset removes all vectors from the index.
//	Reset() error
//
//	// RemoveIDs removes the vectors specified by sel from the index.
//	// Returns the number of elements removed and error.
//	RemoveIDs(sel *faiss.IDSelector) (int, error)
//
//	// Delete frees the memory used by the index.
//	Delete()
//}

type IndexConfiguration struct {
	Embedder llm.Embedder
}

// IndexEntry represents a single entry in the index.
type IndexEntry struct {
	// DocumentID is the unique ID for the document.
	DocumentID string
	// IndexID is the unique ID for the index.
	IndexID int
	// ChunkIndex is the index of the chunk in the document.
	ChunkIndex int
	// ChunkCount is the total number of chunks in the document.
	ChunkCount int
	// Embedding is the embedding for the chunk.
	Embedding []float32
}

type SearchHit struct {
	IndexEntry

	Distance float32
}

// Document represents a single document.
type Document interface {
	ID() string
	Content() string
}

// Index is an index for semantic search.
// It is a wrapper around a FAISS index that allows for indexing and querying.
// It's designed to be used with an embedder like OpenAI's AdaV2.
type Index struct {
	embedder llm.Embedder
	chunker  chunkers.Chunker
	idx      faiss.Index
}

func NewIndex(config IndexConfiguration) (*Index, error) {
	// Create a new index with the given configuration.
	// The dimensionality of the index is determined by the embedder.
	if config.Embedder == nil {
		return nil, fmt.Errorf("Embedder cannot be nil")
	}
	idx, err := faiss.NewIndexFlatIP(1536)

	if err != nil {
		return nil, fmt.Errorf("Failed to create index: %w", err)
	}
	index := &Index{
		embedder: config.Embedder,
		idx:      idx,
	}
	return index, nil
}

// Add adds a document to the index.
func (i *Index) Add(document Document) ([]IndexEntry, error) {
	ctx := context.TODO()

	// Extract the chunks from the document's content.
	chunks, err := i.chunker.SplitTextIntoChunks(ctx, document.Content(), i.embedder.MaxTokensPerChunk(), 32)
	if err != nil {
		return nil, err
	}

	// Get the embeddings for each chunk.
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}
	embeddings, err := i.embedder.GetEmbeddings(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}

	var entries []IndexEntry
	for j, embedding := range embeddings {
		// Get the embedding for the current chunk.
		float32Embedding := embedding.Float32()

		// Add the embedding to the index.
		if err := i.idx.Add(float32Embedding); err != nil {
			return nil, fmt.Errorf("failed to add embedding to index: %w", err)
		}

		// Create an IndexEntry for the current chunk.
		entry := IndexEntry{
			DocumentID: document.ID(),
			IndexID:    j,
			ChunkIndex: j,
			ChunkCount: len(embeddings),
			Embedding:  float32Embedding,
		}

		// Append the entry to the result.
		entries = append(entries, entry)
	}

	return entries, nil
}

// QueryClosestHits returns the closest hits for the given query.
// It searches the index with the query, and returns the top k entries, where each entry consists of an index ID and a distance.
func (i *Index) QueryClosestHits(query string, k int64) ([]SearchHit, error) {
	ctx := context.TODO()
	// Get the embedding for the query.
	embedding, err := i.embedder.GetEmbeddings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}

	return i.QueryClosestHitsWithEmbedding(embedding[0], k)
}

// QueryClosestHitsWithEmbedding returns the closest hits for the given query.
func (i *Index) QueryClosestHitsWithEmbedding(query llm.Embedding, k int64) ([]SearchHit, error) {
	f32embeddings := query.Float32()

	// Search the index for the top k hits.
	// The result is a combined list of distances and indices.
	distances, indices, err := i.idx.Search(f32embeddings, k)
	if err != nil {
		return nil, fmt.Errorf("failed to search index: %w", err)
	}
	// Map index IDs and distances to index entries.
	entries := make([]SearchHit, len(indices))

	for j, index := range indices {
		entries[j] = SearchHit{
			IndexEntry: IndexEntry{
				// TODO: FIXME: Look up the document ID and other fields from the index.
				DocumentID: "",
				IndexID:    int(index),
				ChunkIndex: -1,
				ChunkCount: -1,
				Embedding:  nil,
			},

			Distance: distances[j],
		}
	}

	return entries, nil
}
