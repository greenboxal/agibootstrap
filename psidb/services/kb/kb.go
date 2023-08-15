package kb

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/services/indexing"
)

type KnowledgeRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`

	CurrentDepth int `json:"current_depth"`
	MaxDepth     int `json:"max_depth"`

	References []psi.Path `json:"references"`
	BackLinkTo psi.Path   `json:"back_link_to"`
}

type IKnowledgeBase interface {
	CreateKnowledge(ctx context.Context, node psi.Node, request *KnowledgeRequest) (*Document, error)
}

type KnowledgeBase struct {
	psi.NodeBase

	Name string `json:"name"`

	IndexManager *indexing.Manager `json:"-" inject:""`
}

var KnowledgeBaseInterface = psi.DefineNodeInterface[IKnowledgeBase]()
var KnowledgeBaseType = psi.DefineNodeType[*KnowledgeBase](
	psi.WithInterfaceFromNode(KnowledgeBaseInterface),
)

var EdgeKindKnowledgeBase = psi.DefineEdgeType[*KnowledgeBase]("kb.root")
var EdgeKindKnowledgeBaseDocuments = psi.DefineEdgeType[*indexing.Scope]("kb.documents")

func NewKnowledgeBase() *KnowledgeBase {
	kb := &KnowledgeBase{}
	kb.Init(kb)

	return kb
}

func (kb *KnowledgeBase) PsiNodeName() string { return kb.Name }

func (kb *KnowledgeBase) GetGlobalDocumentScope(ctx context.Context) *indexing.Scope {
	return psi.MustResolveChildOrCreate[*indexing.Scope](
		ctx,
		kb,
		EdgeKindKnowledgeBaseDocuments.Singleton().AsPathElement(),
		func() *indexing.Scope {
			scp := indexing.NewScope()
			scp.SetParent(kb)
			kb.SetEdge(EdgeKindKnowledgeBaseDocuments.Singleton(), scp)
			return scp
		},
	)
}

func (kb *KnowledgeBase) DispatchCreateKnowledge(ctx context.Context, requestor psi.Path, req *KnowledgeRequest) error {
	tx := coreapi.GetTransaction(ctx)

	return tx.Notify(ctx, psi.Notification{
		Notifier:  requestor,
		Notified:  kb.CanonicalPath(),
		Interface: KnowledgeBaseInterface.Name(),
		Action:    "CreateKnowledge",
		Argument:  req,
	})
}

func (kb *KnowledgeBase) CreateKnowledge(ctx context.Context, request *KnowledgeRequest) (*Document, error) {
	tx := coreapi.GetTransaction(ctx)
	doc, err := kb.deduplicateDocument(ctx, request)

	if err != nil {
		return nil, err
	}

	if err := doc.Update(ctx); err != nil {
		return nil, err
	}

	if !request.BackLinkTo.IsEmpty() {
		backLinkTo, err := psi.Resolve[*Document](ctx, tx.Graph(), request.BackLinkTo)

		if err != nil {
			return nil, err
		}

		backLinkTo.SetEdge(EdgeKindRelatedDocument.Named(slugify(request.Title)), doc)

		if err := backLinkTo.Update(ctx); err != nil {
			return nil, err
		}
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

func (kb *KnowledgeBase) ResolveCategory(ctx context.Context, name string) (*Category, error) {
	tx := coreapi.GetTransaction(ctx)
	catPath := kb.CanonicalPath().Child(psi.PathElement{Name: name})

	return psi.ResolveOrCreate[*Category](ctx, tx.Graph(), catPath, func() *Category {
		cat := NewCategory(slugify(name))
		cat.SetEdge(EdgeKindKnowledgeBase.Named("root"), kb)

		return cat
	})
}

func (kb *KnowledgeBase) deduplicateDocument(ctx context.Context, request *KnowledgeRequest) (*Document, error) {
	tx := coreapi.GetTransaction(ctx)
	slug := slugify(request.Title)
	docPath := kb.CanonicalPath().Child(psi.PathElement{Name: slug})

	existing, err := psi.Resolve[*Document](ctx, tx.Graph(), docPath)

	if err != nil && !errors.Is(err, psi.ErrNodeNotFound) {
		return nil, err
	}

	if existing != nil {
		return existing, nil
	}

	idx, err := kb.GetGlobalDocumentScope(ctx).GetIndex(ctx)

	if err != nil {
		return nil, err
	}

	q, err := idx.Embedder().EmbeddingsForNode(ctx, stdlib.NewText(fmt.Sprintf("# %s\n%s\n", request.Title, request.Description)))

	if err != nil {
		return nil, err
	}

	for q.Next() {
		hits, err := idx.Search(ctx, indexing.SearchRequest{
			Graph:      tx.Graph(),
			Query:      q.Value(),
			Limit:      5,
			ReturnNode: true,
		})

		if err != nil {
			return nil, err
		}

		docs := iterators.Map(hits, func(t indexing.NodeSearchHit) *Document {
			return t.Node.(*Document)
		})

		for docs.Next() {
			dedup, err := QueryDocumentDeduplication(ctx, request, docs.Value())

			if err != nil {
				return nil, err
			}

			isDup := dedup.IsDuplicated && (dedup.IsSameContent && dedup.IsSamePerspective && dedup.IsSameTopic && dedup.IsSameSubject) || (dedup.IsSameHeading && dedup.IsSynonym)

			if isDup && dedup.Index > 0 {
				return kb.deduplicateDocument(ctx, &KnowledgeRequest{
					Title:        dedup.Title,
					Description:  request.Description,
					References:   request.References,
					CurrentDepth: request.CurrentDepth,
					MaxDepth:     request.MaxDepth,
				})
			}
		}
	}

	doc, err := psi.ResolveOrCreate[*Document](ctx, tx.Graph(), docPath, func() *Document {
		doc := NewDocument()
		doc.Title = request.Title
		doc.Description = request.Description
		doc.Slug = slugify(doc.Title)
		doc.SetParent(kb)
		doc.SetEdge(EdgeKindKnowledgeBase.Named("root"), kb)

		return doc
	})

	if err != nil {
		return nil, err
	}

	return doc, nil
}
