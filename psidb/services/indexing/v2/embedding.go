package indexing

import (
	"context"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Embedding struct {
	psi.NodeBase

	Origin psi.Path    `json:"O"`
	V      [][]float32 `json:"V"`
}

var EmbeddingType = psi.DefineNodeType[*Embedding]()
var EmbeddingEdge = psi.DefineEdgeType[*Embedding]("gpt.embedding")

func (e *Embedding) PsiNodeName() string { return "_Embedding" }

func NewEmbedding() *Embedding {
	e := &Embedding{}
	e.Init(e, psi.WithNodeType(EmbeddingType))

	return e
}

func ResolveEmbeddings(ctx context.Context, node psi.Node) (*Embedding, error) {
	tx := coreapi.GetTransaction(ctx)
	path := node.CanonicalPath().Child(psi.PathElement{Name: "_Embedding"})

	embedding, err := psi.Resolve[*Embedding](ctx, tx.Graph(), path)

	if err != nil {
		return nil, err
	}

	return embedding, nil
}
