package docs

import (
	"strconv"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type IndexEntry struct {
	psi.NodeBase

	Key   string   `json:"K"`
	Query psi.Path `json:"Q"`
}

func NewIndexEntry(key string, query psi.Path) *IndexEntry {
	ie := &IndexEntry{
		Key:   key,
		Query: query,
	}

	ie.Init(ie)

	return ie
}

func (ie *IndexEntry) PsiNodeName() string { return ie.Key }

func (ie *IndexEntry) GetAllChunks() []*IndexEntryChunk {
	return psi.GetEdges[*IndexEntryChunk](ie, IndexEntryChunkEdge)
}

func (ie *IndexEntry) GetChunk(index int) *IndexEntryChunk {
	return psi.GetEdgeOrNil[*IndexEntryChunk](ie, IndexEntryChunkEdge.Named(strconv.Itoa(index)))
}

func (ie *IndexEntry) AddChunk(chunk *IndexEntryChunk) {
	chunk.SetParent(ie)
	ie.SetEdge(IndexEntryChunkEdge.Named(strconv.Itoa(chunk.ChunkIndex)), chunk)
}

func (ie *IndexEntry) RemoveChunk(index int) {
	chunk := ie.GetChunk(index)

	if chunk == nil {
		return
	}

	ie.RemoveChildNode(chunk)
}

type IndexEntryChunk struct {
	psi.NodeBase

	Ordinal    int64     `json:"O"`
	ChunkIndex int       `json:"I"`
	Embeddings []float32 `json:"E"`
	Content    string    `json:"C"`
}

func NewIndexEntryChunk(ord int64, index int, embeddings []float32, content string) *IndexEntryChunk {
	iec := &IndexEntryChunk{
		Ordinal:    ord,
		ChunkIndex: index,
		Embeddings: embeddings,
		Content:    content,
	}

	iec.Init(iec)

	return iec
}

func (iec *IndexEntryChunk) PsiNodeName() string { return strconv.FormatInt(iec.Ordinal, 10) }

var IndexEntryType = psi.DefineNodeType[*IndexEntry]()
var IndexEntryChunkType = psi.DefineNodeType[*IndexEntryChunk]()
var IndexEntryEdge = psi.DefineEdgeType[*IndexEntry]("psidb.docs.indexentry")
var IndexEntryChunkEdge = psi.DefineEdgeType[*IndexEntryChunk]("IEC")
