package kb

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type KnowledgeRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`

	CurrentDepth int `json:"current_depth"`
	MaxDepth     int `json:"max_depth"`

	References []psi.Path `json:"references"`
}

type IKnowledgeBase interface {
	CreateKnowledge(ctx context.Context, node psi.Node, request *KnowledgeRequest) (*Document, error)
}

type KnowledgeBase struct {
	psi.NodeBase

	Name string `json:"name"`
}

var KnowledgeBaseInterface = psi.DefineNodeInterface[IKnowledgeBase]()
var KnowledgeBaseType = psi.DefineNodeType[*KnowledgeBase](
	psi.WithInterfaceFromNode(KnowledgeBaseInterface),
)

var EdgeKindKnowledgeBase = psi.DefineEdgeType[*KnowledgeBase]("kb.root")

func NewKnowledgeBase() *KnowledgeBase {
	kb := &KnowledgeBase{}
	kb.Init(kb)

	return kb
}

func (kb *KnowledgeBase) PsiNodeName() string { return kb.Name }

func (kb *KnowledgeBase) CreateKnowledge(ctx context.Context, request *KnowledgeRequest) (*Document, error) {
	tx := coreapi.GetTransaction(ctx)
	slug := slugify(request.Title)
	docPath := kb.CanonicalPath().Child(psi.PathElement{Name: slug})

	doc, err := psi.ResolveOrCreate[*Document](ctx, tx.Graph(), docPath, func() *Document {
		doc := NewDocument()
		doc.Title = request.Title
		doc.Description = request.Description
		doc.Slug = slugify(doc.Title)
		doc.SetParent(kb)

		return doc
	})

	if err != nil {
		return nil, err
	}

	doc.SetEdge(EdgeKindKnowledgeBase.Named("root"), kb)

	if err := doc.Update(ctx); err != nil {
		return nil, err
	}

	if err := doc.DispatchLearn(ctx, kb.CanonicalPath(), &LearnRequest{
		References:   request.References,
		CurrentDepth: request.CurrentDepth,
		MaxDepth:     request.MaxDepth,
	}); err != nil {
		return nil, err
	}

	return doc, nil
}
