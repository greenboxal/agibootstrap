package gpt

import (
	"bytes"
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/chunkers"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering/themes"
	indexing2 "github.com/greenboxal/agibootstrap/psidb/services/indexing"
)

var Module = fx.Module(
	"modules/gpt",

	fx.Provide(NewEmbeddingCacheManager),
	fx.Provide(NewDefaultNodeEmbedder),

	fx.Invoke(func(sp inject.ServiceProvider, e indexing2.NodeEmbedder) {
		inject.RegisterInstance(sp, e)
	}),
)

func NewDefaultNodeEmbedder() indexing2.NodeEmbedder {
	return &DefaultEmbedder{
		chunker: chunkers.TikToken{},
	}
}

type DefaultEmbedder struct {
	chunker chunkers.Chunker
}

func (d *DefaultEmbedder) Dimensions() int {
	return 1536
}

func (d *DefaultEmbedder) EmbeddingsForNode(ctx context.Context, n psi.Node) (indexing2.GraphEmbeddingIterator, error) {
	var buffer bytes.Buffer

	err := rendering.RenderNodeWithTheme(ctx, &buffer, themes.GlobalTheme, "text/markdown", "", n)

	if err != nil {
		return nil, err
	}

	texts, err := d.chunker.SplitTextIntoStrings(ctx, buffer.String(), 256, 0)

	if err != nil {
		return nil, err
	}

	chunks, err := gpt.GlobalEmbedder.GetEmbeddings(ctx, texts)

	if err != nil {
		return nil, err
	}

	return iterators.Map(iterators.FromSlice(chunks), func(v llm.Embedding) indexing2.GraphEmbedding {
		return indexing2.GraphEmbedding{
			Semantic: v.Embeddings,
		}
	}), nil
}
