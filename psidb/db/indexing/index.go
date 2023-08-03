package indexing

import (
	"context"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type NodeSearchHit struct {
	BasicSearchHit

	Node psi.Node
}

type GraphEmbeddingIterator = iterators.Iterator[GraphEmbedding]

type NodeEmbedder interface {
	Dimensions() int

	EmbeddingsForNode(ctx context.Context, n psi.Node) (GraphEmbeddingIterator, error)
}

type NodeIndex interface {
	Index() BasicIndex

	Embedder() NodeEmbedder

	IndexNode(ctx context.Context, n psi.Node) error
	Search(ctx context.Context, req SearchRequest) (iterators.Iterator[NodeSearchHit], error)

	Close() error
}

type nodeIndex struct {
	manager  *Manager
	embedder NodeEmbedder
	index    BasicIndex
}

func (ni *nodeIndex) Index() BasicIndex {
	return ni.index
}

func (ni *nodeIndex) Embedder() NodeEmbedder {
	return ni.embedder
}

func (ni *nodeIndex) IndexNode(ctx context.Context, n psi.Node) error {
	var link *cidlink.Link

	path := n.CanonicalPath()

	if snap := n.PsiNodeBase().GetSnapshot(); snap != nil && snap.CommitLink() != nil {
		l := snap.CommitLink().(cidlink.Link)
		link = &l
	}

	embeddings, err := ni.embedder.EmbeddingsForNode(ctx, n)

	if err != nil {
		return err
	}

	for i := 0; embeddings.Next(); i++ {
		embedding := embeddings.Value()

		_, err = ni.index.IndexNode(ctx, IndexNodeRequest{
			Path:       &path,
			Link:       link,
			ChunkIndex: int64(i),
			Embeddings: embedding,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (ni *nodeIndex) Search(ctx context.Context, req SearchRequest) (iterators.Iterator[NodeSearchHit], error) {
	basicHits, err := ni.index.Search(ctx, req)

	if err != nil {
		return nil, err
	}

	return iterators.NewIterator(func() (NodeSearchHit, bool) {
		for {
			if !basicHits.Next() {
				return NodeSearchHit{}, false
			}

			hit := NodeSearchHit{
				BasicSearchHit: basicHits.Value(),
			}

			if hit.Path != nil {
				//hit.Node, err = ni.graph.ResolveNode(ctx, *hit.Path)
				panic("fix me")
			}

			if err != nil {
				ni.manager.logger.Warn("failed to resolve node", "err", err)
				continue
			}

			if hit.Node == nil {
				ni.manager.logger.Warn("failed to resolve node", "err", err)
				continue
			}

			return hit, true
		}
	}), nil
}

func (ni *nodeIndex) Close() error {
	return ni.index.Close()
}
