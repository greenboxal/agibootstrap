package kb

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"gonum.org/v1/gonum/floats"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/services/indexing"
)

type ConnectionLink struct {
	From []psi.Path `json:"from"`
	To   []psi.Path `json:"to"`
}
type Connection struct {
	psi.NodeBase

	// Name of the connection
	Name string `json:"name" jsonschema:"title=Name of the connection,description=The unique identifier for the connection"`

	// Root path of the connection
	Root psi.Path `json:"root" jsonschema:"title=Root path of the connection,description=The root path for the connection"`

	// From path of the connection
	From psi.Path `json:"from" jsonschema:"title=From path of the connection,description=The starting path for the connection"`

	// To path of the connection
	To psi.Path `json:"to" jsonschema:"title=To path of the connection,description=The destination path for the connection"`

	// Frontier paths of the connection
	Frontier []psi.Path `json:"frontier" jsonschema:"title=Frontier paths of the connection,description=The frontier paths for the connection"`

	// GScore of the connection
	GScore map[string]float64 `json:"g_score" jsonschema:"title=GScore of the connection,description=The GScore for the connection"`

	// FScore of the connection
	FScore map[string]float64 `json:"f_score" jsonschema:"title=FScore of the connection,description=The FScore for the connection"`

	// Links of the connection

	Links map[string][]psi.Path `json:"links" jsonschema:"title=Links of the connection,description=The links for the connection"`

	// Embedder of the connection
	Embedder indexing.NodeEmbedder `inject:"" json:"-" jsonschema:"title=Embedder of the connection,description=The embedder for the connection"`
}

func (conn *Connection) PsiNodeName() string { return conn.Name }

type IConnection interface {
	Step(ctx context.Context) error
	ProcessEdge(ctx context.Context, req *ConnectionEdgeRequest) error
	Reset(ctx context.Context) error
}

var ConnectionInterface = psi.DefineNodeInterface[IConnection]()
var ConnectionType = psi.DefineNodeType[*Connection](
	psi.WithInterfaceFromNode(ConnectionInterface),
)

func NewConnection() *Connection {
	c := &Connection{}
	c.Init(c, psi.WithNodeType(ConnectionType))

	return c
}

func (conn *Connection) Step(ctx context.Context) error {
	var frontier *psi.Path

	frontierIndex := -1

	for i, p := range conn.Frontier {
		if frontier == nil || conn.FScore[p.String()] < conn.FScore[frontier.String()] {
			frontier = &p
			frontierIndex = i
		}
	}

	if frontier == nil {
		return nil
	}

	conn.Frontier = slices.Delete(conn.Frontier, frontierIndex, frontierIndex+1)

	tx := coreapi.GetTransaction(ctx)
	current, err := psi.Resolve[*Document](ctx, tx.Graph(), *frontier)

	if err != nil {
		return err
	}

	if current.CanonicalPath().Equals(conn.To) {
		return nil
	}

	edgeCount := 0
	promise := tx.MakePromise()

	for edges := current.Edges(); edges.Next(); {
		edge := edges.Value()
		neighbor, ok := edge.To().(*Document)

		if !ok {
			continue
		}

		if err := conn.dispatchEdge(ctx, *frontier, neighbor.CanonicalPath(), promise.Signal(1)); err != nil {
			return err
		}

		edgeCount++
	}

	nextStep := psi.Notification{
		Notifier:  conn.CanonicalPath(),
		Notified:  conn.CanonicalPath(),
		Interface: "IConnection",
		Action:    "Step",
		Argument:  struct{}{},
	}

	if edgeCount > 0 {
		nextStep.Dependencies = append(nextStep.Dependencies, promise.Wait(edgeCount))
	} else if edgeCount == 0 {
		learnReq := &LearnRequest{
			CurrentDepth: 0,
			MaxDepth:     1,
			Observer:     promise.Signal(1),
		}

		if err != nil {
			return err
		}

		if err := current.DispatchLearn(ctx, conn.CanonicalPath(), learnReq); err != nil {
			return err
		}

		conn.Frontier = append(conn.Frontier, *frontier)

		nextStep.Dependencies = append(nextStep.Dependencies, promise.Wait(1))
	}

	if err := tx.Notify(ctx, nextStep); err != nil {
		return err
	}

	return conn.Update(ctx)
}

func (conn *Connection) Reset(ctx context.Context) error {
	conn.FScore = make(map[string]float64)
	conn.GScore = make(map[string]float64)
	conn.Frontier = []psi.Path{conn.From}
	conn.Links = make(map[string][]psi.Path)

	fromKey := conn.From.String()

	if h, err := conn.calculateHeuristic(ctx, conn.From); err != nil {
		return err
	} else {
		conn.GScore[fromKey] = 0
		conn.FScore[fromKey] = h
	}

	return conn.Update(ctx)
}

func (conn *Connection) calculateDistance(ctx context.Context, a, b psi.Path) (float64, error) {
	ae, err := conn.getEmbedding(ctx, a)

	if err != nil {
		return 0, err
	}

	be, err := conn.getEmbedding(ctx, b)

	if err != nil {
		return 0, err
	}

	return floats.Dot(ae, be), nil
}

func (conn *Connection) calculateHeuristic(ctx context.Context, from psi.Path) (float64, error) {
	return conn.calculateDistance(ctx, from, conn.To)
}

func (conn *Connection) getEmbedding(ctx context.Context, p psi.Path) ([]float64, error) {
	tx := coreapi.GetTransaction(ctx)
	n, err := tx.Resolve(ctx, p)

	if err != nil {
		return nil, err
	}

	embeddings, err := conn.Embedder.EmbeddingsForNode(ctx, n)

	if err != nil {
		return nil, err
	}

	if !embeddings.Next() {
		return nil, nil
	}

	return embeddings.Value().ToFloat64Slice(nil), nil
}

type ConnectionEdgeRequest struct {
	Frontier psi.Path `json:"frontier"`
	To       psi.Path `json:"to"`

	Expand bool `json:"expand"`

	Observer psi.Promise `json:"observers,omitempty"`
}

func (conn *Connection) ProcessEdge(ctx context.Context, req *ConnectionEdgeRequest) error {
	npath := req.To
	nkey := npath.String()
	fkey := req.Frontier.String()

	tx := coreapi.GetTransaction(ctx)
	current, err := psi.Resolve[*Document](ctx, tx.Graph(), req.Frontier)

	if !current.HasContent {

		promise := tx.MakePromise()

		learnReq := &LearnRequest{
			CurrentDepth: 0,
			MaxDepth:     1,
			Observer:     promise.Signal(1),
		}

		if err != nil {
			return err
		}

		if err := current.DispatchLearn(ctx, conn.CanonicalPath(), learnReq); err != nil {
			return err
		}

		return tx.Notify(ctx, psi.Notification{
			Notifier: conn.CanonicalPath(),
			Notified: conn.CanonicalPath(),

			Interface: "IConnection",
			Action:    "ProcessEdge",

			Argument: &ConnectionEdgeRequest{
				Frontier: req.Frontier,
				To:       req.To,
				Observer: req.Observer,
				Expand:   false,
			},

			Dependencies: []psi.Promise{promise.Wait(1)},
		})
	}

	d, err := conn.calculateDistance(ctx, req.Frontier, req.To)

	if err != nil {
		return err
	}

	tentativeGScore := conn.GScore[fkey] + d

	if score, ok := conn.GScore[nkey]; !ok || tentativeGScore < score {
		h, err := conn.calculateHeuristic(ctx, npath)

		if err != nil {
			return err
		}

		doc, err := conn.Diffuse(ctx, req.Observer, req.Frontier, npath)

		if err != nil {
			return err
		}

		conn.GScore[nkey] = tentativeGScore
		conn.FScore[nkey] = tentativeGScore + h
		conn.Links[nkey] = append(conn.Links[nkey], doc.CanonicalPath())
		conn.Frontier = append(conn.Frontier, doc.CanonicalPath())
	}

	return conn.Update(ctx)
}

func (conn *Connection) Diffuse(ctx context.Context, promise psi.Promise, frontierPath, nextPath psi.Path) (*Document, error) {
	tx := coreapi.GetTransaction(ctx)

	frontier, err := psi.Resolve[*Document](ctx, tx.Graph(), frontierPath)

	if err != nil {
		return nil, err
	}

	next, err := psi.Resolve[*Document](ctx, tx.Graph(), nextPath)

	if err != nil && !errors.Is(err, psi.ErrNodeNotFound) {
		return nil, err
	}

	goal, err := psi.Resolve[*Document](ctx, tx.Graph(), conn.To)

	if err != nil && !errors.Is(err, psi.ErrNodeNotFound) {
		return nil, err
	}

	title := frontier.Title + " and " + next.Title + " in the context of " + goal.Title
	slug := slugify(title)
	p := conn.Root.Child(psi.PathElement{Name: slug})

	doc, err := psi.ResolveOrCreate[*Document](ctx, tx.Graph(), p, func() *Document {
		doc := NewDocument()
		doc.Title = title
		doc.Description = title
		doc.Slug = slugify(doc.Title)
		doc.Root = conn.Root

		return doc
	})

	if err != nil {
		return nil, err
	}

	if !doc.HasContent {
		if err := doc.DispatchLearn(ctx, conn.CanonicalPath(), &LearnRequest{
			CurrentDepth: 0,
			MaxDepth:     1,
			Observer:     promise.Signal(1),
		}); err != nil {
			return nil, err
		}

		if err := tx.Wait(ctx, promise.Wait(1)); err != nil {
			return nil, err
		}
	}

	if err := doc.Update(ctx); err != nil {
		return nil, err
	}

	return doc, nil
}

func (conn *Connection) dispatchEdge(ctx context.Context, frontier, to psi.Path, onComplete psi.Promise) error {
	tx := coreapi.GetTransaction(ctx)

	return tx.Notify(ctx, psi.Notification{
		Notifier: conn.CanonicalPath(),
		Notified: conn.CanonicalPath(),

		Interface: "IConnection",
		Action:    "ProcessEdge",

		Argument: &ConnectionEdgeRequest{
			Frontier: frontier,
			To:       to,
			Observer: onComplete,
			Expand:   true,
		},
	})
}
