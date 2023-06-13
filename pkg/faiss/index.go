package faiss

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"sort"

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

// Add adds an IndexEntry to the Index. The parameter item should contain an embedding.
// This method fills item.IndexID and adds item.Embedding to the index. It also adds the item to the cache
// using a hash of its ID as the key and returns an error if the embedding is nil or has an invalid dimension.
func (i *Index) Add(item IndexEntry) error {
	if item.Embedding == nil {
		return errors.New("embedding is nil")
	}
	if len(item.Embedding) != i.dimension {
		return fmt.Errorf("invalid dimension for embedding. expected: %d, got: %d", i.dimension, len(item.Embedding))
	}

	i.totalItems++
	item.IndexID = i.totalItems - 1

	if err := i.cache.Set(hash(item.ID), item.Embedding, 1); err != nil {
		i.logger.Errorw("error setting embedding to cache", "item_id", item.ID)
	}

	item.ChunkIndex = 0
	item.ChunkCount = 1

	if i.verbose {
		i.logger.Infow("adding item to index", "index_id", item.IndexID, "item_id", item.ID)
	}
	if err := i.index.Add(item.Embedding); err != nil {
		return fmt.Errorf("error adding embedding %v to index with index_id: %d, chunk_index: %d, chunk_count: %d. %s", item.Embedding, item.IndexID, item.ChunkIndex, item.ChunkCount, err.Error())
	}

	return nil
}

// hash is a utility function that takes a string and returns a 64-bit hash.
//
// Parameters:
// s (string): the string to be hashed.
//
// Returns:
// uint64: a 64-bit hash.
func hash(s string) uint64 {
	h := fnv.New64a()
	if _, err := h.Write([]byte(s)); err != nil {
		panic(err)
	}
	return h.Sum64()
}

func (i *Index) QueryClosestHits(query Embedding, k int) ([]*IndexEntry, error) {
	// if query embeddings size is zero, return nil
	if len(query.Embeddings) == 0 {
		return nil, nil
	}

	// if query embeddings size is not equal to index dimension, return nil
	if len(query.Embeddings) != i.dimension {
		return nil, nil
	}

	// create slices for index distances and indices
	dists := make([]float32, i.totalItems)
	idx := make([]int, i.totalItems)

	// search the index for the closest items to the query
	_, _, err := i.index.Search(query.Float32(), k, dists, idx)
	if err != nil {
		return nil, err
	}

	// sort indices based on their distances
	sort.Slice(idx, func(a, b int) bool {
		return dists[a] < dists[b]
	})

	// create a slice to store k-closest items to the query
	var res []*IndexEntry

	// for each of the k-closest indices, fetch the item from the cache and append its details to the result slice
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

	return res, nil
}

func indexToId(indexID int, keyPrefix string) string {
	return fmt.Sprintf("%s_%d", keyPrefix, indexID)
}

func hashIndex(indexID int, keyPrefix string) uint64 {
	return hash(indexToId(indexID, keyPrefix))
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
