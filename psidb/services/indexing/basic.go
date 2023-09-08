package indexing

import (
	"context"
	"fmt"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type IndexedItem struct {
	Index      int64    `json:"index"`
	Path       psi.Path `json:"path"`
	ChunkIndex int64    `json:"chunkIndex"`

	Embeddings *GraphEmbedding `json:"embeddings,omitempty"`
}

func (i IndexedItem) Identity() string {
	return fmt.Sprintf("%s\000%d", i.Path, i.ChunkIndex)
}

type SearchRequest struct {
	Graph coreapi.GraphOperations
	Query GraphEmbedding
	Limit int

	ReturnEmbeddings bool
	ReturnNode       bool
}

type BasicSearchHit struct {
	IndexedItem `json:"indexedItem"`

	Score float32 `json:"score,omitempty"`
}

type IndexNodeRequest struct {
	Path       psi.Path       `json:"path"`
	ChunkIndex int64          `json:"chunkIndex"`
	Embeddings GraphEmbedding `json:"embeddings,omitempty"`
}

type BasicIndex interface {
	Dimensions() int

	IndexNode(ctx context.Context, req IndexNodeRequest) (IndexedItem, error)
	Search(ctx context.Context, req SearchRequest) (iterators.Iterator[BasicSearchHit], error)

	Rebuild(ctx context.Context) error
	Truncate(ctx context.Context) error

	Load() error
	Save() error

	Close() error
}
