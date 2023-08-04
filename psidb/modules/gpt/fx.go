package gpt

import (
	"bytes"
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering/themes"
	"github.com/greenboxal/agibootstrap/psidb/db/indexing"
)

var Module = fx.Module(
	"modules/gpt",

	fx.Provide(NewEmbeddingCacheManager),
	fx.Provide(NewDefaultNodeEmbedder),

	fx.Invoke(func(sp inject.ServiceProvider, e indexing.NodeEmbedder) {
		inject.RegisterInstance(sp, e)
	}),
)

func NewDefaultNodeEmbedder() indexing.NodeEmbedder {
	return &DefaultEmbedder{}
}

type DefaultEmbedder struct {
}

func (d *DefaultEmbedder) Dimensions() int {
	return 1536
}

func (d *DefaultEmbedder) EmbeddingsForNode(ctx context.Context, n psi.Node) (indexing.GraphEmbeddingIterator, error) {
	var buffer bytes.Buffer

	err := rendering.RenderNodeWithTheme(ctx, &buffer, themes.GlobalTheme, "text/markdown", "", n)

	if err != nil {
		return nil, err
	}

	texts := []string{buffer.String()}

	chunks, err := gpt.GlobalEmbedder.GetEmbeddings(ctx, texts)

	if err != nil {
		return nil, err
	}

	return iterators.Map(iterators.FromSlice(chunks), func(v llm.Embedding) indexing.GraphEmbedding {
		return indexing.GraphEmbedding{
			Semantic: v.Embeddings,
		}
	}), nil
}
