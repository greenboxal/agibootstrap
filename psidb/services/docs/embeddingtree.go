package docs

import (
	"context"
	"math"

	"github.com/greenboxal/aip/aip-langchain/pkg/chunkers"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"
	"golang.org/x/exp/slices"
	"gonum.org/v1/gonum/floats"
)

type EmbeddingTreeChunk struct {
	Index   int
	Tokens  tokenizers.Tokens
	Content string

	Embedding  []float64
	ChunkDelta []float64
}

type EmbeddingTreeLevel struct {
	Level int
	Width int

	Chunks []*EmbeddingTreeChunk
}

type EmbeddingTreeOverlay struct {
	Chunk *EmbeddingTreeChunk

	T int
	L int
	X float64

	DeltaEmbedding []float64
}

func (eto *EmbeddingTreeOverlay) FromChunk(l int, chunk *EmbeddingTreeChunk) {
	eto.Chunk = chunk

	eto.L = l
	eto.X = float64(chunk.Index) * (1.0 / float64(int(1)<<l))
}

type EmbeddingTree struct {
	embedder  llm.Embedder
	tokenizer tokenizers.BasicTokenizer

	maxLevels      int
	causalitySpeed int

	levels  []*EmbeddingTreeLevel
	overlay []*EmbeddingTreeOverlay
}

func NewEmbeddingTree(
	tokenizer tokenizers.BasicTokenizer,
	embedder llm.Embedder,
	causalitySpeed int,
	data []byte,
) *EmbeddingTree {
	tokens, err := tokenizer.GetTokens(data)

	if err != nil {
		panic(err)
	}

	et := &EmbeddingTree{
		causalitySpeed: causalitySpeed,
		maxLevels:      int(math.Ceil(math.Log2(float64(tokens.Len())))),

		tokenizer: tokenizer,
		embedder:  embedder,
	}

	et.levels = make([]*EmbeddingTreeLevel, et.maxLevels)

	et.levels[0] = &EmbeddingTreeLevel{
		Chunks: []*EmbeddingTreeChunk{
			{
				Index:   0,
				Tokens:  tokens,
				Content: tokens.String(),
			},
		},
	}

	return et
}

func (et *EmbeddingTree) Build(ctx context.Context) error {
	for i := 0; i < et.maxLevels; i++ {
		err := et.buildLevel(ctx, i)

		if err != nil {
			return err
		}
	}

	return et.buildOverlay(ctx)
}

func (et *EmbeddingTree) buildLevel(ctx context.Context, level int) error {
	root := et.levels[0].Chunks[0]
	target := et.levels[level]

	if level > 0 {
		target = &EmbeddingTreeLevel{}
		et.levels[level] = target

		target.Width = et.levels[level-1].Width / 2
	} else {
		target.Width = root.Tokens.Len()
	}

	overlapSize := target.Width / 2
	chunkSize := target.Width - overlapSize

	if target.Width < 32 {
		target.Width = 0
		return nil
	}

	// Calculate the number of chunks based on the chunk size and overlap size
	numChunks := (root.Tokens.Len() + chunkSize - 1) / chunkSize

	target.Level = level

	if level > 0 {
		target.Chunks = make([]*EmbeddingTreeChunk, 0, numChunks)

		// Generate the chunks by iterating over the tokens
		for i := 0; i < numChunks; i++ {
			startIndex := i * chunkSize
			endIndex := chunkers.Min(startIndex+target.Width, root.Tokens.Len())

			if endIndex-startIndex <= 0 {
				break
			}

			chunkTokens := root.Tokens.Slice(startIndex, endIndex)

			target.Chunks = append(target.Chunks, &EmbeddingTreeChunk{
				Index:   i,
				Tokens:  chunkTokens,
				Content: chunkTokens.String(),
			})
		}
	}

	if target.Width <= et.embedder.MaxTokensPerChunk() {
		strs := make([]string, len(target.Chunks))

		for i, chunk := range target.Chunks {
			strs[i] = chunk.Content
		}

		embeddings, err := et.embedder.GetEmbeddings(ctx, strs)

		if err != nil {
			return err
		}

		for i, e := range embeddings {
			target.Chunks[i].Embedding = e.Float64()
		}
	}

	return nil
}

func (et *EmbeddingTree) buildOverlay(ctx context.Context) error {
	et.overlay = make([]*EmbeddingTreeOverlay, 0)

	for _, level := range et.levels {
		if level.Width == 0 {
			continue
		}

		eto := &EmbeddingTreeOverlay{}
		eto.FromChunk(level.Level, level.Chunks[0])

		et.overlay = append(et.overlay, eto)
	}

	slices.SortFunc(et.overlay, func(i, j *EmbeddingTreeOverlay) bool {
		return i.X < j.X
	})

	runningEmbedding := make([]float64, et.embedder.Dimensions())

	for i, eto := range et.overlay {
		eto.T = i

		floats.Add(runningEmbedding, eto.Chunk.Embedding)
		eto.DeltaEmbedding = slices.Clone(runningEmbedding)
	}

	return nil
}
