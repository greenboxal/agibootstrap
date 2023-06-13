package faiss

import (
	"context"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/dgraph-io/ristretto"
	"go.uber.org/zap"
)

// IndexConfiguration is the configuration for an Index.
type IndexConfiguration struct {
	Embedder Embedder
}

// IndexEntry is an entry in an Index.
// TODO: Write documentation for this.
type IndexEntry struct {
	ID         string
	IndexID    int
	ChunkIndex int
	ChunkCount int
	Embedding  []float32
	Document   Document
}

// Document
// TODO: Write documentation for this.
type Document interface {
	Content() string
}

// Index
// TODO: Write documentation for this.
type Index struct {
	index      faiss.Index
	dimension  int
	batchSize  int
	codeSize   int
	nlist      int
	verbose    bool
	totalItems int
	keyPrefix  string
	cache      *ristretto.Cache
	logger     *zap.SugaredLogger
}

// NewIndex function creates a new Faiss index with the configurations passed in the IndexConfiguration struct.
// It returns a pointer to an instance of the Index struct.
//
// Parameters:
// config (IndexConfiguration): The configuration object for the Faiss index.
//
// Returns:
// (*Index): A pointer to an instance of the Index struct.
func NewIndex(config IndexConfiguration) *Index {
	dimension := config.Embedder.Dim()
	batchSize := 1024
	codeSize := 512
	nlist := 64
	totalItems := 0
	keyPrefix := "faiss_index"
	lfuCacheStartSize := 1024
	lfuCacheMaxSize := 131072
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(lfuCacheStartSize * 10),
		MaxCost:     int64(lfuCacheMaxSize),
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	index, err := faiss.NewIndexFlatIP(dimension)
	if err != nil {
		panic(err)
	}
	return &Index{
		index:      index,
		dimension:  dimension,
		batchSize:  batchSize,
		codeSize:   codeSize,
		nlist:      nlist,
		totalItems: totalItems,
		keyPrefix:  keyPrefix,
		cache:      cache,
	}
}

func (i *Index) Add(item IndexEntry) error {
	return nil
}

func (i *Index) QueryClosestHits(query Embedding, k int) ([]*IndexEntry, error) {
	return nil, nil
}

type Embedder interface {
	Dim() int
	MaxTokensPerChunk() int

	GetEmbeddings(ctx context.Context, chunks []string) ([]Embedding, error)
}

func (e Embedding) Float64() []float64 {
	return e.Embeddings
}

func (e Embedding) Float32() []float32 {
	arr := make([]float32, len(e.Embeddings))
	for i, v := range e.Embeddings {
		arr[i] = float32(v)
	}
	return arr
}

func (e Embedding) Dim() int {
	return len(e.Embeddings)
}

type Embedding struct {
	Embeddings []float64
}
