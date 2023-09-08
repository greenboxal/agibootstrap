package kb

import (
	"context"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
	"gonum.org/v1/gonum/floats"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/indexing"
)

type KnowledgeRequest struct {
	Title       string `json:"title" jsonschema:"title=Title of the document,description=Title of the document"`
	Description string `json:"description" jsonschema:"title=Description of the document,description=Description of the document"`

	CurrentDepth int `json:"current_depth" jsonschema:"title=Current depth,description=Current depth"`
	MaxDepth     int `json:"max_depth" jsonschema:"title=Max depth,description=Max depth"`

	References []psi.Path                   `json:"references" jsonschema:"title=References,description=References"`
	BackLinkTo *stdlib.Reference[*Document] `json:"back_link_to" jsonschema:"title=Back link to,description=Back link to document"`

	Observer psi.Promise `json:"observer" jsonschema:"title=Observer,description=Observer promise"`
}

type TraceRequest struct {
	From psi.Path `json:"from"`
	To   psi.Path `json:"to"`

	Dispatch bool `json:"dispatch"`
}

type TraceResponse struct {
	Trace []psi.Path `json:"trace"`
}

type IKnowledgeBase interface {
	CreateKnowledge(ctx context.Context, request *KnowledgeRequest) (*Document, error)
	TraceConcept(ctx context.Context, request *TraceRequest) (*TraceResponse, error)
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

func (kb *KnowledgeBase) TraceConcept(ctx context.Context, request *TraceRequest) (*TraceResponse, error) {
	var result TraceResponse

	if !request.Dispatch {
		tx := coreapi.GetTransaction(ctx)
		err := tx.Notify(ctx, psi.Notification{
			Notifier:  kb.CanonicalPath(),
			Notified:  kb.CanonicalPath(),
			Interface: KnowledgeBaseInterface.Name(),
			Action:    "TraceConcept",
			Argument: &TraceRequest{
				From:     request.From,
				To:       request.To,
				Dispatch: true,
			},
		})

		if err != nil {
			return nil, err
		}

		return &result, nil
	}

	index := kb.GetGlobalDocumentScope(ctx)
	embedder := index.Embedder

	from, err := psi.Resolve[*Document](ctx, coreapi.GetTransaction(ctx).Graph(), request.From)

	if err != nil {
		return nil, err
	}

	to, err := psi.Resolve[*Document](ctx, coreapi.GetTransaction(ctx).Graph(), request.To)

	if err != nil {
		return nil, err
	}

	openSet := []psi.Node{from}
	cameFrom := map[psi.Node]psi.Node{}
	gScore := map[psi.Node]float64{}
	fScore := map[psi.Node]float64{}
	eCache := map[psi.Node][]float64{}

	rebuildPath := func(n psi.Node) []psi.Node {
		path := []psi.Node{n}

		for {
			if prev, ok := cameFrom[n]; ok {
				path = append(path, prev)
				n = prev
			} else {
				break
			}
		}

		return path
	}

	getEmbedding := func(n psi.Node) []float64 {
		if e, ok := eCache[n]; ok {
			return e
		}

		e, err := embedder.EmbeddingsForNode(ctx, n)

		if err != nil {
			panic(err)
		}

		if !e.Next() {
			panic("no embeddings")
		}

		arr := e.Value().ToFloat64Slice(nil)

		eCache[n] = arr

		return arr
	}

	calculateDistance := func(a, b psi.Node) float64 {
		ae := getEmbedding(a)
		be := getEmbedding(b)

		return floats.Dot(ae, be)
	}

	calculateHeuristic := func(a psi.Node) float64 {
		return calculateDistance(a, to)
	}

	gScore[from] = 0
	fScore[from] = calculateHeuristic(from)

	for len(openSet) > 0 {
		slices.SortFunc(openSet, func(i, j psi.Node) bool {
			return calculateHeuristic(i) < calculateHeuristic(j)
		})

		current := openSet[0].(*Document)
		openSet = openSet[1:]

		if current.CanonicalPath().Equals(request.To) {
			p := rebuildPath(current)

			result.Trace = lo.Map(p, func(n psi.Node, _ int) psi.Path { return n.CanonicalPath() })

			break
		}

		gotEdges := false

		for !gotEdges {
			learnReq := &LearnRequest{
				CurrentDepth: 0,
				MaxDepth:     2,
			}

			if !current.HasContent {
				if err := current.Learn(ctx, learnReq); err != nil {
					return nil, err
				}
			} else {
				if err := current.Expand(ctx, learnReq); err != nil {
					return nil, err
				}
			}

			if err != nil {
				return nil, err
			}

			for edges := current.Edges(); edges.Next(); {
				edge := edges.Value()
				neighbor, ok := edge.To().(*Document)

				if !ok {
					continue
				}

				gotEdges = true

				tentativeGScore := gScore[current] + calculateDistance(current, neighbor)

				if score, ok := gScore[neighbor]; !ok || tentativeGScore < score {
					cameFrom[neighbor] = current
					gScore[neighbor] = tentativeGScore
					fScore[neighbor] = tentativeGScore + calculateHeuristic(neighbor)

					openSet = append(openSet, neighbor)
				}
			}
		}
	}

	return &result, nil
}

func (kb *KnowledgeBase) CreateKnowledge(ctx context.Context, request *KnowledgeRequest) (*Document, error) {
	doc, err := kb.deduplicateDocument(ctx, request)

	if err != nil {
		return nil, err
	}

	if err := doc.Update(ctx); err != nil {
		return nil, err
	}

	if !request.BackLinkTo.IsEmpty() {
		backLinkTo, err := request.BackLinkTo.Resolve(ctx)

		if err != nil {
			return nil, err
		}

		backLinkTo.SetEdge(EdgeKindRelatedDocument.Named(slugify(request.Title)), doc)

		if err := backLinkTo.Update(ctx); err != nil {
			return nil, err
		}
	}

	learnReq := &LearnRequest{
		References:   request.References,
		CurrentDepth: request.CurrentDepth,
		MaxDepth:     request.MaxDepth,
		Observer:     request.Observer,
	}

	if err := doc.DispatchLearn(ctx, kb.CanonicalPath(), learnReq); err != nil {
		return nil, err
	}

	return doc, doc.Update(ctx)
}

func (kb *KnowledgeBase) ResolveCategory(ctx context.Context, name string) (*Category, error) {
	tx := coreapi.GetTransaction(ctx)
	catPath := kb.CanonicalPath().Child(psi.PathElement{Name: name})

	return psi.ResolveOrCreate[*Category](ctx, tx.Graph(), catPath, func() *Category {
		cat := NewCategory(slugify(name))

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

	/*idx, err := kb.GetGlobalDocumentScope(ctx).GetIndex(ctx)

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
	}*/

	doc, err := psi.ResolveOrCreate[*Document](ctx, tx.Graph(), docPath, func() *Document {
		doc := NewDocument()
		doc.Title = request.Title
		doc.Description = request.Description
		doc.Slug = slugify(doc.Title)
		doc.Root = kb.CanonicalPath()
		doc.SetParent(kb)

		return doc
	})

	if err != nil {
		return nil, err
	}

	return doc, nil
}
