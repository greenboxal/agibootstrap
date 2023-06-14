package codex

import "context"

type Index interface {
	Add(ctx context.Context, document Document) ([]IndexEntry, error)
	Search(ctx context.Context, query string, k int) ([]SearchHit, error)
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

type Indexer struct {
	project *Project
	index   Index
}

func (p *Indexer) Index() error {
	// TODO: Implement
	// Implement indexing logic here
	return nil
}
