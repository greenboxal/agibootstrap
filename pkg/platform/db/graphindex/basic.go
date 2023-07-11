package graphindex

import (
	"context"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type IndexedItem struct {
	Index      int64 `json:"index"`
	ChunkIndex int64 `json:"chunkIndex"`

	Path *psi.Path     `json:"path"`
	Link *cidlink.Link `json:"link"`

	Embeddings GraphEmbedding `json:"embeddings,omitempty"`
}

type SearchRequest struct {
	Query GraphEmbedding
	Limit int
}

type BasicSearchHit struct {
	IndexedItem `json:"indexedItem"`

	Score float32 `json:"score,omitempty"`
}

type IndexNodeRequest struct {
	Path       *psi.Path     `json:"path"`
	Link       *cidlink.Link `json:"link"`
	ChunkIndex int64         `json:"chunkIndex"`

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
