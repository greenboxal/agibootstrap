package faiss

import (
	"context"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/dgraph-io/ristretto"
	"go.uber.org/zap"
)

// TODO: Add documentation

type IndexConfiguration struct {
	Embedder Embedder
}

type IndexEntry struct {
	ID         string
	IndexID    int
	ChunkIndex int
	ChunkCount int
	Embedding  []float32
	Document   Document
}

type Document interface {
	Content() string
}

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
	// It should fill item.Index and add item.Embedding to the index
	if item.Embedding == nil {
		return errors.New("embedding is nil")
	}
	if len(item.Embedding) != i.dimension {
		return fmt.Errorf("invalid dimension for embedding. expected: %d, got: %d", i.dimension, len(item.Embedding))
	}

	i.totalItems++
	index := i.totalItems - 1
	item.IndexID = index

	if err := i.cache.Set(hash(item.ID), item.Embedding, 1); err != nil {
		i.logger.Errorw("error setting embedding to cache", "item_id", item.ID)
	}

	item.ChunkIndex = 0
	item.ChunkCount = 1
	// Add item to index
	if i.verbose {
		i.logger.Infow("adding item to index", "index_id", index, "item_id", item.ID)
	}
	if _, err := i.index.Add(item.Embedding); err != nil {
		return fmt.Errorf("error adding embedding %+v to index with index_id: %d, chunk_index: %d, chunk_count: %d. %s", item.Embedding, index, item.ChunkIndex, item.ChunkCount, err.Error())
	}
	if i.totalItems%i.batchSize == 0 {
		return i.trainIndex()
	}

	return nil
}

func hash(s string) uint64 {
	return fnv.New64a().Sum64([]byte(s))
}

func (i *Index) QueryClosestHits(query Embedding, k int) []*IndexEntry {

	if len(query.Embeddings) == 0 {
		return nil
	}

	if len(query.Embeddings) != i.dimension {
		return nil
	}

	dists := make([]float32, i.totalItems)
	idx := make([]int, i.totalItems)

	i.index.Search(query.Float32(), k, dists, idx)

	sort.Slice(idx, func(a, b int) bool {
		return dists[a] < dists[b]
	})

	var res []*IndexEntry
	for _, j := range idx[:k] {
		val, ok := i.cache.Get(hashIndex(j))
		if ok {
			res = append(res, &IndexEntry{
				ID:         indexToId(j),
				Embedding:  val.([]float32),
				IndexID:    j,
				ChunkIndex: 0,
				ChunkCount: 0,
			})
		}
	}

	return res
}

func indexToId(indexID int) string {
	return fmt.Sprintf("%s_%d", i.keyPrefix, indexID)
}

func hashIndex(indexID int) uint64 {
	return fnv.New64a().Sum64([]byte(indexToId(indexID)))
}

func (e Embedding) Float32() []float32 {
	arr := make([]float32, len(e.Embeddings))
	for i, v := range e.Embeddings {
		arr[i] = float32(v)
	}
	return arr
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
