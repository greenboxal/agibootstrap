package indexing

import (
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
)

type Index[K comparable] interface {
	ReadOnlyIndex[K]

	Add(key K, value ...llm.Embedding) error
	Remove(key K) bool
}

type ReadOnlyIndex[K comparable] interface {
	Query(q llm.Embedding, k int64) ([]SearchHit[K], error)
}

// IndexEntry represents a single entry in the index.
type IndexEntry[K comparable] struct {
	// DocumentID is the unique ID for the document.
	DocumentID K
	// IndexID is the unique ID for the index.
	IndexID int64
	// ChunkIndex is the index of the chunk in the document.
	ChunkIndex int
	// ChunkCount is the total number of chunks in the document.
	ChunkCount int
	// Embedding is the embedding for the chunk.
	Embedding []float32
	// Valid indicates whether the entry is valid or not.
	Valid bool
}

// SearchHit represents a single search hit.
type SearchHit[K comparable] struct {
	IndexEntry[K]

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
