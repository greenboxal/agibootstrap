package indexing

import (
	"bytes"
	"context"
	"fmt"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type IndexedItem struct {
	Index int64 `json:"index"`

	Path *psi.Path     `json:"path"`
	Link *cidlink.Link `json:"link"`

	ChunkIndex int64         `json:"chunkIndex"`
	ChunkLink  *cidlink.Link `json:"chunkLink"`

	Embeddings GraphEmbedding `json:"embeddings,omitempty"`
}

func (i IndexedItem) Identity() string {
	w := bytes.NewBuffer(nil)

	fmt.Fprintf(w, "%d\000", i.Index)
	fmt.Fprintf(w, "%s\000", i.Path)
	fmt.Fprintf(w, "%s\000", i.Link)
	fmt.Fprintf(w, "%d\000", i.ChunkIndex)
	fmt.Fprintf(w, "%s\000", i.ChunkLink)
	fmt.Fprintf(w, "%f\000", i.Embeddings.ToFloat32Slice(nil))

	mh, err := multihash.Sum(w.Bytes(), multihash.SHA2_256, -1)

	if err != nil {
		panic(err)
	}

	return mh.B58String()
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
	Path *psi.Path     `json:"path"`
	Link *cidlink.Link `json:"link"`

	ChunkIndex int64         `json:"chunkIndex"`
	ChunkLink  *cidlink.Link `json:"chunkLink"`

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
