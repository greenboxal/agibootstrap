package kb

import (
	"context"
	"fmt"

	"github.com/gomarkdown/markdown/html"
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type LearnRequest struct {
	CurrentDepth int        `json:"current_depth" jsonschema:"title=Current Depth,description=The current depth of the learning request"`
	MaxDepth     int        `json:"max_depth" jsonschema:"title=Max Depth,description=The maximum depth of the learning request"`
	Feedback     string     `json:"feedback" jsonschema:"title=Feedback,description=The feedback for the learning request"`
	References   []psi.Path `json:"references" jsonschema:"title=References,description=The references for the learning request"`

	Observer psi.Promise `json:"observers,omitempty" jsonschema:"title=Observer,description=The observer for the learning request"`
}

// IDocument is an interface that defines the methods a Document should implement.
type IDocument interface {
	// Learn is a method that takes a context and a LearnRequest as parameters and returns an error.
	// It is responsible for learning from the provided request.
	Learn(ctx context.Context, req *LearnRequest) error

	// Expand is a method that takes a context and a LearnRequest as parameters and returns an error.
	// It is responsible for expanding the knowledge based on the provided request.
	Expand(ctx context.Context, req *LearnRequest) error
}

type Document struct {
	psi.NodeBase

	Root psi.Path `json:"root" jsonschema:"title=Root,description=The root path of the document"`

	Slug        string `json:"slug" jsonschema:"title=Slug,description=The slug of the document"`
	Title       string `json:"title" jsonschema:"title=Title,description=The title of the document"`
	Description string `json:"description" jsonschema:"title=Description,description=The description of the document"`

	Categories    []string `json:"categories" jsonschema:"title=Categories,description=The categories of the document"`
	RelatedTopics []string `json:"related_topics" jsonschema:"title=Related Topics,description=The related topics of the document"`

	Body    string `json:"body" jsonschema:"title=Body,description=The body of the document"`
	Summary string `json:"summary" jsonschema:"title=Summary,description=The summary of the document"`

	HasContent bool `json:"has_content" jsonschema:"title=Has Content,description=Indicates if the document has content"`
	HasSummary bool `json:"has_summary" jsonschema:"title=Has Summary,description=Indicates if the document has a summary"`
}

var DocumentInterface = psi.DefineNodeInterface[IDocument]()
var DocumentType = psi.DefineNodeType[*Document](
	psi.WithInterfaceFromNode(DocumentInterface),
)

var EdgeKindRelatedDocument = psi.DefineEdgeType[*Document]("related")

func NewDocument() *Document {
	d := &Document{}
	d.Init(d)

	return d
}

func (d *Document) PsiNodeName() string { return d.Slug }

func (d *Document) Learn(ctx context.Context, req *LearnRequest) error {
	var err error

	tx := coreapi.GetTransaction(ctx)
	kb := psi.MustResolve[*KnowledgeBase](ctx, tx.Graph(), d.Root)

	if !d.HasContent {
		if err := d.generateContent(ctx, req); err != nil {
			return nil
		}

		d.HasContent = true
	}

	if !d.HasSummary && d.HasContent {
		if err := d.generateSummary(ctx, req); err != nil {
			return nil
		}

		d.HasSummary = true
	}

	if err := d.Update(ctx); err != nil {
		return err
	}

	scp := kb.GetGlobalDocumentScope(ctx)
	idx, err := scp.GetIndex(ctx)

	if err != nil {
		return err
	}

	if err := idx.IndexNode(ctx, d); err != nil {
		return err
	}

	return d.Expand(ctx, req)
}

func (d *Document) generateSummary(ctx context.Context, req *LearnRequest) error {
	var msgs []*thoughtdb.Thought

	msgs = append(msgs, &thoughtdb.Thought{
		From: thoughtdb.CommHandle{Role: msn.RoleSystem},
		Text: `You are an AI assistant specialized in summarizing documents. You are given a document to summarize.`,
	})

	if req.Feedback != "" {
		msgs = append(msgs, &thoughtdb.Thought{
			From: thoughtdb.CommHandle{Role: msn.RoleUser},
			Text: fmt.Sprintf("**Feedback:**\n%s\n", req.Feedback),
		})
	}

	msgs = append(msgs, &thoughtdb.Thought{
		From: thoughtdb.CommHandle{Role: msn.RoleUser},
		Text: fmt.Sprintf("\n%s\n%s\n", d.Title, d.Body),
	})

	msgs = append(msgs, &thoughtdb.Thought{
		From: thoughtdb.CommHandle{Role: msn.RoleUser},
		Text: "Summarize the document above.",
	})

	reply, err := runChainWithMessages(ctx, msgs)

	if err != nil {
		return err
	}

	d.Summary = reply

	return nil
}

func (d *Document) generateContent(ctx context.Context, req *LearnRequest) error {
	var msgs []*thoughtdb.Thought

	tx := coreapi.GetTransaction(ctx)

	references := make([]*Document, 0, len(req.References))
	for _, ref := range req.References {
		doc, err := psi.Resolve[*Document](ctx, tx.Graph(), ref)

		if err != nil {
			return err
		}

		references = append(references, doc)
	}

	msgs = append(msgs, &thoughtdb.Thought{
		From: thoughtdb.CommHandle{Role: msn.RoleSystem},
		Text: `You are an AI assistant specialized in writing documents. You are given a document to write.`,
	})

	if req.Feedback != "" {
		msgs = append(msgs, &thoughtdb.Thought{
			From: thoughtdb.CommHandle{Role: msn.RoleUser},
			Text: fmt.Sprintf("**Feedback:**\n%s\n", req.Feedback),
		})
	}

	for _, ref := range references {
		msgs = append(msgs, &thoughtdb.Thought{
			From: thoughtdb.CommHandle{Role: msn.RoleUser},
			Text: fmt.Sprintf("\n%s\n%s\n", ref.Title, ref.Body),
		})
	}

	if d.HasContent {
		msgs = append(msgs, &thoughtdb.Thought{
			From: thoughtdb.CommHandle{Role: msn.RoleUser},
			Text: fmt.Sprintf("%s\n%s\n", d.Title, d.Body),
		})

		msgs = append(msgs, &thoughtdb.Thought{
			From: thoughtdb.CommHandle{Role: msn.RoleUser},
			Text: "Write a better, newer version of the document above.",
		})
	} else {
		msgs = append(msgs, &thoughtdb.Thought{
			From: thoughtdb.CommHandle{Role: msn.RoleUser},
			Text: fmt.Sprintf("%s\n%s\n", d.Title, d.Description),
		})

		msgs = append(msgs, &thoughtdb.Thought{
			From: thoughtdb.CommHandle{Role: msn.RoleUser},
			Text: fmt.Sprintf("Write an article about '%s'.", d.Title),
		})
	}

	reply, err := runChainWithMessages(ctx, msgs)

	if err != nil {
		return err
	}

	d.Body = reply

	return nil
}

func (d *Document) Categorize(ctx context.Context, req *LearnRequest) error {
	var history []*thoughtdb.Thought

	if req.CurrentDepth >= req.MaxDepth {
		return nil
	}

	tx := coreapi.GetTransaction(ctx)
	kb := psi.MustResolve[*KnowledgeBase](ctx, tx.Graph(), d.Root)

	doct := thoughtdb.NewThought()
	doct.From.Role = msn.RoleUser
	doct.Text = fmt.Sprintf("# %s\n%s\n", d.Title, d.Body)
	history = append(history, doct)

	res, err := QueryDocumentCategories(ctx, history)

	if err != nil {
		return err
	}

	for _, categoryName := range res.Categories {
		cat, err := kb.ResolveCategory(ctx, categoryName)

		if err != nil {
			return err
		}

		cat.AddDocument(d)
	}

	return d.Update(ctx)
}

func (d *Document) Expand(ctx context.Context, req *LearnRequest) error {
	var history []*thoughtdb.Thought

	tx := coreapi.GetTransaction(ctx)
	kb := psi.MustResolve[*KnowledgeBase](ctx, tx.Graph(), d.Root)

	doct := thoughtdb.NewThought()
	doct.From.Role = msn.RoleUser
	doct.Text = fmt.Sprintf("# %s\n%s\n", d.Title, d.Body)
	history = append(history, doct)

	res, err := QueryDocumentRelatedContent(ctx, history)

	if err != nil {
		return err
	}

	for _, entry := range res.Related {
		kbReq := &KnowledgeRequest{
			Title:       entry.Title,
			Description: entry.Description,

			CurrentDepth: req.CurrentDepth + 1,
			MaxDepth:     req.MaxDepth,

			BackLinkTo: d.CanonicalPath(),

			Observer: req.Observer.Signal(1),
		}

		err := kb.DispatchCreateKnowledge(ctx, d.CanonicalPath(), kbReq)

		if err != nil {
			return err
		}

		if err := tx.Wait(ctx, req.Observer.Wait(1)); err != nil {
			return err
		}
	}

	if err := tx.Signal(ctx, req.Observer); err != nil {
		return err
	}

	return d.Update(ctx)
}

func (d *Document) DispatchLearn(ctx context.Context, requestor psi.Path, req *LearnRequest) error {
	tx := coreapi.GetTransaction(ctx)

	req.CurrentDepth++

	if req.CurrentDepth > req.MaxDepth {
		return tx.Signal(ctx, req.Observer)
	}

	logger.Infow("Dispatching learn request", "requestor", requestor, "notified", d.CanonicalPath())

	return tx.Notify(ctx, psi.Notification{
		Notifier:  requestor,
		Notified:  d.CanonicalPath(),
		Interface: DocumentInterface.Name(),
		Action:    "Learn",
		Argument:  req,
	})
}

func (d *Document) DispatchExpand(ctx context.Context, requestor psi.Path, req *LearnRequest) error {
	tx := coreapi.GetTransaction(ctx)

	logger.Infow("Dispatching expand request", "requestor", requestor, "notified", d.CanonicalPath())

	return tx.Notify(ctx, psi.Notification{
		Notifier:  requestor,
		Notified:  d.CanonicalPath(),
		Interface: DocumentInterface.Name(),
		Action:    "Expand",
		Argument:  req,
	})
}

func slugify(s string) string {
	return string(html.Slugify([]byte(s)))
}
