package graphindex

import (
	"bytes"
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/chunkers"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering/themes"
)

type AnchoredEmbedder struct {
	Base    llm.Embedder
	Chunker chunkers.Chunker

	Root   psi.Node
	Anchor psi.Node
}

func (a *AnchoredEmbedder) Dimensions() int {
	if a.Base == nil {
		return 8
	}

	return a.Base.Dimensions() + 5
}

func (a *AnchoredEmbedder) EmbeddingsForNode(ctx context.Context, n psi.Node) (GraphEmbeddingIterator, error) {
	baseEmbedding := GraphEmbedding{}
	baseEmbedding.Depth = n.CanonicalPath().Depth()
	baseEmbedding.TreeDistance = a.calculateTreeDistance(a.Anchor, n)
	baseEmbedding.ReferenceDistance = a.calculateReferenceDistance(a.Anchor, n)
	baseEmbedding.Time = a.calculateTimeMetric(a.Anchor, n)

	buffer := bytes.NewBuffer(nil)

	if err := rendering.RenderNodeWithTheme(buffer, themes.GlobalTheme, "text/markdown", "", n); err != nil {
		return nil, err
	}

	chunks, err := a.Chunker.SplitTextIntoStrings(ctx, buffer.String(), a.Base.MaxTokensPerChunk(), a.Base.MaxTokensPerChunk()/10)

	if err != nil {
		return nil, err
	}

	embeddings, err := a.Base.GetEmbeddings(ctx, chunks)

	if err != nil {
		return nil, err
	}

	embeddingIterator := iterators.FromSlice(embeddings)

	return iterators.NewIterator(func() (ge GraphEmbedding, ok bool) {
		if !embeddingIterator.Next() {
			return ge, false
		}

		ge = baseEmbedding
		ge.Semantic = embeddingIterator.Value().Embeddings

		return ge, true
	}), nil
}

func (a *AnchoredEmbedder) calculateTimeMetric(anchor psi.Node, n psi.Node) int64 {
	return 0
}

func (a *AnchoredEmbedder) calculateTreeDistance(anchor psi.Node, target psi.Node) int {
	rootPath := a.Root.CanonicalPath()
	anchorPath := anchor.CanonicalPath()
	targetPath := target.CanonicalPath()

	commonRoot, ok := anchorPath.GetCommonAncestor(targetPath)

	if !ok {
		return -1
	}

	if !rootPath.IsAncestorOf(commonRoot) {
		return -1
	}

	anchorDistance := anchorPath.Depth() - rootPath.Depth()
	targetDistance := targetPath.Depth() - rootPath.Depth()
	commonDistance := commonRoot.Depth() - rootPath.Depth()

	return anchorDistance + targetDistance - commonDistance
}

func (a *AnchoredEmbedder) calculateReferenceDistance(anchor psi.Node, n psi.Node) int {
	return 0
}
