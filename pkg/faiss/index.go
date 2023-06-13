package faiss

import "C"
import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

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

type Key []string

type ObjectStore interface {
	Get(ctx context.Context, key Key) ([]byte, error)
	Put(ctx context.Context, key Key, value []byte) error
}

type IndexConfiguration struct {
	// Name is the name of the index.
	Name string
	// Embedder is the embedder to use for the index.
	Embedder llm.Embedder
	// Chunker is the chunker to use for the index.
	Chunker chunkers.Chunker
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

// SearchHit represents a single search hit.
type SearchHit struct {
	IndexEntry

	// Distance is the distance between the query and the document.
	Distance float32
}

// Document represents a single document.
type Document interface {
	// ID returns the unique ID for the document.
	ID() string

	// Content returns the content of the document.
	Content() string
}

// Index is an index for semantic search.
// It is a wrapper around a FAISS index that allows for indexing and querying.
// It's designed to be used with an embedder like OpenAI's AdaV2.
type Index struct {
	embedder    llm.Embedder
	chunker     chunkers.Chunker
	faiss       faiss.Index
	objectStore ObjectStore
	name        string
}

// NewIndex creates a new index with the given configuration.
// The dimensionality of the index is determined by the embedder.
func NewIndex(objectStore ObjectStore, config IndexConfiguration) (*Index, error) {
	if config.Embedder == nil {
		return nil, fmt.Errorf("embedder cannot be nil")
	}

	if config.Chunker == nil {
		return nil, fmt.Errorf("chunker cannot be nil")
	}

	idx, err := faiss.NewIndexFlatIP(1536)

	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	index := &Index{
		name:        config.Name,
		embedder:    config.Embedder,
		chunker:     config.Chunker,
		objectStore: objectStore,
		faiss:       idx,
	}
	return index, nil
}

// Add adds a document to the index.
func (idx *Index) Add(ctx context.Context, document Document) ([]IndexEntry, error) {
	// Extract the chunks from the document's content.
	chunks, err := idx.chunker.SplitTextIntoChunks(ctx, document.Content(), idx.embedder.MaxTokensPerChunk(), 32)
	if err != nil {
		return nil, err
	}

	// Get the embeddings for each chunk.
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}
	embeddings, err := idx.embedder.GetEmbeddings(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}

	var entries []IndexEntry
	for j, embedding := range embeddings {
		// Get the embedding for the current chunk.
		float32Embedding := embedding.Float32()

		// Add the embedding to the index.
		if err := idx.faiss.Add(float32Embedding); err != nil {
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

		// Use the ObjectStore to store the embedding.
		for j, embedding := range embeddings {
			float32Embedding := embedding.Float32()
			entry := IndexEntry{
				DocumentID: document.ID(),
				IndexID:    j,
				ChunkIndex: j,
				ChunkCount: len(embeddings),
				Embedding:  float32Embedding,
			}

			data, err := json.Marshal(entry)

			if err != nil {
				return nil, fmt.Errorf("failed to marshal entry: %w", err)
			}

			key := Key{"faiss", idx.name, strconv.Itoa(j)}

			// Store the embedding with the given key in the object store
			err = idx.objectStore.Put(ctx, key, data)

			if err != nil {
				return nil, fmt.Errorf("failed to store embedding with key %v: %w", key, err)
			}

			entries = append(entries, entry)
		}

		// Append the entry to the result.
		entries = append(entries, entry)
	}

	return entries, nil
}

// QueryClosestHits returns the closest hits for the given query.
// It searches the index with the query, and returns the top k entries, where each entry consists of an index ID and a distance.
func (idx *Index) QueryClosestHits(ctx context.Context, query string, k int64) ([]SearchHit, error) {
	// Get the embedding for the query.
	embedding, err := idx.embedder.GetEmbeddings(ctx, []string{query})

	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}

	return idx.QueryClosestHitsWithEmbedding(ctx, embedding[0], k)
}

// QueryClosestHitsWithEmbedding returns the closest hits for the given query.
func (idx *Index) QueryClosestHitsWithEmbedding(ctx context.Context, query llm.Embedding, k int64) ([]SearchHit, error) {
	f32embeddings := query.Float32()

	// Search the index for the top k hits.
	// The result is a combined list of distances and indices.
	distances, indices, err := idx.faiss.Search(f32embeddings, k)
	if err != nil {
		return nil, fmt.Errorf("failed to search index: %w", err)
	}
	// Map index IDs and distances to index entries.
	entries := make([]SearchHit, len(indices))

	for j, index := range indices {
		// Look up the document ID and other fields from the index.
		entry, err := idx.GetEntryByID(ctx, index)

		if err != nil {
			return nil, fmt.Errorf("failed to get index entry: %w", err)
		}

		entries[j] = SearchHit{
			IndexEntry: entry,

			Distance: distances[j],
		}
	}

	return entries, nil
}

// GetEntryByID Get the index entry with the given ID.
func (idx *Index) GetEntryByID(ctx context.Context, id int64) (IndexEntry, error) {
	entry := IndexEntry{
		DocumentID: "",
		IndexID:    int(id),
		ChunkIndex: 0,
		ChunkCount: 0,
		Embedding:  nil,
	}

	key := Key{"faiss", idx.name, strconv.FormatInt(id, 10)}

	// Store the embedding with the given key in the object store
	data, err := idx.objectStore.Get(ctx, key)

	if err != nil {
		return entry, fmt.Errorf("failed to store embedding with key %v: %w", key, err)
	}

	if err := json.Unmarshal(data, &entry); err != nil {
		return entry, fmt.Errorf("failed to unmarshal entry: %w", err)
	}

	return entry, nil
}
