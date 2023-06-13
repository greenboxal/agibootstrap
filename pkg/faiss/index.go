package faiss

import (
	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/dgraph-io/ristretto"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"go.uber.org/zap"
	"gonum.org/v1/gonum/mat"
)

type IndexConfiguration struct {
	Embedder llm.Embedder
}

type Index struct {
	index      *faiss.Index
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
	verbose := true
	totalItems := 0
	keyPrefix := "faiss_index"
	lfuCacheStartSize := 1024
	lfuCacheMaxSize := 131072
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: lfuCacheStartSize * 10,
		MaxCost:     int64(lfuCacheMaxSize),
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	logger := logrus.New()
	index := faiss.NewIndexFlatIP(codeSize, dimension)
	index.SetNumProbes(10)
	return &Index{
		index:      index,
		dimension:  dimension,
		batchSize:  batchSize,
		codeSize:   codeSize,
		nlist:      nlist,
		verbose:    verbose,
		totalItems: totalItems,
		keyPrefix:  keyPrefix,
		cache:      cache,
		logger:     logger,
	}
}

func (i *Index) trainIndex(ctx context.Context) error {
	// Train a faiss index
	return nil
}

func (e Embedding) Vector() *mat.VecDense {
	return mat64.NewVecDense(len(e.Embeddings), e.Float64())
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
