package docs

import (
	"bytes"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenboxal/aip/aip-langchain/pkg/chunkers"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering/themes"
)

type AddNodeRequest struct {
	Node *stdlib.Reference[psi.Node] `json:"node"`
}

type RemoveNodeRequest struct {
	Node *stdlib.Reference[psi.Node] `json:"node"`
}

type SearchNodesRequest struct {
	QueryPrompt string                      `json:"query_prompt,omitempty"`
	QueryNode   *stdlib.Reference[psi.Node] `json:"query_node,omitempty"`
	Limit       int                         `json:"limit"`
}

type SearchNodeHit struct {
	Entry *IndexEntry      `json:"entry"`
	Chunk *IndexEntryChunk `json:"chunk"`
	Score float32          `json:"score"`
}

type SearchNodeResponse struct {
	Hits []*SearchNodeHit `json:"hits"`
}

type IIndex interface {
	AddNode(ctx context.Context, req *AddNodeRequest) error
	RemoveNode(ctx context.Context, req *RemoveNodeRequest) error

	SearchNodes(ctx context.Context, req *SearchNodesRequest) (*SearchNodeResponse, error)
}

type Index struct {
	psi.NodeBase

	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	Counter int32  `json:"counter"`

	IndexManager *IndexManager `json:"-" inject:""`
	Index        *LiveIndex    `json:"-"`

	EmbeddingManager *gpt.EmbeddingCacheManager `json:"-" inject:""`
	Embedder         llm.Embedder               `json:"-"`

	Client *openai.Client `json:"-" inject:""`
}

var IndexInterface = psi.DefineNodeInterface[IIndex]()
var IndexType = psi.DefineNodeType[*Index](psi.WithInterfaceFromNode(IndexInterface))
var _ IIndex = &Index{}

func (idx *Index) Init(self psi.Node) {
	idx.NodeBase.Init(self, psi.WithNodeType(IndexType))

	if idx.UUID == "" {
		idx.UUID = uuid.NewString()
	}
}

func (idx *Index) PsiNodeName() string { return idx.Name }

func (idx *Index) GetEmbedder() llm.Embedder {
	if idx.Embedder == nil {
		idx.Embedder = idx.EmbeddingManager.GetEmbedder(&openai.Embedder{
			Client: idx.Client,
			Model:  openai.AdaEmbeddingV2,
		})
	}

	return idx.Embedder
}

func (idx *Index) GetLiveIndex() *LiveIndex {
	if idx.Index == nil {
		idx.Index = idx.IndexManager.GetOrCreateLiveIndex(idx.UUID)
	}

	return idx.Index
}

func (idx *Index) AddNode(ctx context.Context, req *AddNodeRequest) error {
	node, err := req.Node.Resolve(ctx)

	if err != nil {
		return err
	}

	entry, oldChunks, newChunks, err := idx.prepareNode(ctx, node)

	for _, chunk := range oldChunks {
		err := idx.GetLiveIndex().RemoveChunk(chunk.Ordinal)

		if err != nil {
			return err
		}

		entry.RemoveChunk(chunk.ChunkIndex)
	}

	for _, chunk := range newChunks {
		err := idx.GetLiveIndex().AddChunk(chunk)

		if err != nil {
			return err
		}

		idx.SetEdge(IndexEntryChunkEdge.Indexed(chunk.Ordinal), chunk)
	}

	if entry.Parent() == nil {
		node.SetParent(idx)
	}

	node.SetEdge(IndexEntryEdge.Named(idx.UUID), entry)

	if err := node.Update(ctx); err != nil {
		return err
	}

	return idx.Update(ctx)
}

func (idx *Index) RemoveNode(ctx context.Context, req *RemoveNodeRequest) error {
	node, err := req.Node.Resolve(ctx)

	if err != nil {
		return err
	}

	entryEdge := node.GetEdge(IndexEntryEdge.Named(idx.UUID))

	if entryEdge == nil {
		return nil
	}

	n, err := entryEdge.ResolveTo(ctx)

	if err != nil {
		return err
	}

	entry := n.(*IndexEntry)
	oldChunks := entry.GetAllChunks()

	for _, chunk := range oldChunks {
		if err := idx.GetLiveIndex().RemoveChunk(chunk.Ordinal); err != nil {
			return err
		}
	}

	return nil
}

func (idx *Index) SearchNodes(ctx context.Context, req *SearchNodesRequest) (*SearchNodeResponse, error) {
	var query psi.Node

	tx := coreapi.GetTransaction(ctx)

	if req.QueryNode == nil {
		query = stdlib.NewText(req.QueryPrompt)
	} else {
		q, err := req.QueryNode.Resolve(ctx)

		if err != nil {
			return nil, err
		}

		query = q
	}

	_, queryChunks, err := idx.makeChunks(ctx, query)

	if err != nil {
		return nil, err
	}

	if len(queryChunks) > 1 {
		return nil, errors.New("query node must be a single chunk")
	}

	distances, ords, err := idx.GetLiveIndex().QueryChunks(queryChunks[0], req.Limit)

	if err != nil {
		return nil, err
	}

	hits := make([]*SearchNodeHit, 0, len(ords))

	for i, ord := range ords {
		if ord == -1 {
			continue
		}

		score := distances[i]
		chunk := psi.ResolveEdge(idx, IndexEntryChunkEdge.Indexed(ord))

		if chunk == nil {
			continue
		}

		entry, err := tx.Resolve(ctx, chunk.CanonicalPath().Parent())

		if err != nil {
			continue
		}

		hits = append(hits, &SearchNodeHit{
			Entry: entry.(*IndexEntry),
			Chunk: chunk,
			Score: score,
		})
	}

	return &SearchNodeResponse{
		Hits: hits,
	}, nil
}

func (idx *Index) prepareNode(ctx context.Context, node psi.Node) (entry *IndexEntry, oldChunks, newChunks []*IndexEntryChunk, err error) {
	var key string

	entryEdge := node.GetEdge(IndexEntryEdge.Named(idx.UUID))

	if entryEdge != nil {
		n, err := entryEdge.ResolveTo(ctx)

		if err != nil {
			return nil, nil, nil, err
		}

		entry = n.(*IndexEntry)

		oldChunks = entry.GetAllChunks()
	} else {
		entry = NewIndexEntry("", node.CanonicalPath())
	}

	key, newChunks, err = idx.makeChunks(ctx, node)

	if err != nil {
		return nil, nil, nil, err
	}

	if key == entry.Key {
		return entry, nil, nil, nil
	}

	entry.Key = key
	entry.SetParent(idx)

	if err := entry.Update(ctx); err != nil {
		return nil, nil, nil, err
	}

	for _, chunk := range newChunks {
		entry.AddChunk(chunk)
	}

	return
}

func (idx *Index) makeChunks(ctx context.Context, node psi.Node) (key string, newChunks []*IndexEntryChunk, err error) {
	var buffer bytes.Buffer

	err = rendering.RenderNodeWithTheme(ctx, &buffer, themes.GlobalTheme, "text/markdown", "", node)

	if err != nil {
		return "", nil, err
	}

	mh, err := multihash.Sum(buffer.Bytes(), multihash.SHA2_256, 32)

	if err != nil {
		return "", nil, err
	}

	key = mh.HexString()

	strs, err := chunkers.TikToken{}.SplitTextIntoStrings(ctx, buffer.String(), 8000, 0)

	if err != nil {
		return key, nil, err
	}

	embeddings, err := idx.GetEmbedder().GetEmbeddings(ctx, strs)

	if err != nil {
		return key, nil, err
	}

	for i, str := range strs {
		idx.Counter++

		ord := ((int64(idx.Counter) << 32) & 0x7FFFFFFF) | (time.Now().Unix() & 0xFFFFFFFF)
		chunk := NewIndexEntryChunk(ord, i, embeddings[i].Embeddings, str)

		newChunks = append(newChunks, chunk)
	}

	return
}
