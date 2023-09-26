package docs

import (
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

	"github.com/greenboxal/agibootstrap/psidb/modules/gpt/autoform"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type IDocument interface {
	RegenerateMetadata(ctx context.Context) error
}

type Document struct {
	psi.NodeBase

	UUID    string `json:"uuid"`
	Title   string `json:"title"`
	Content string `json:"content"`

	Observations string `json:"observations"`
}

func (d *Document) PsiNodeName() string { return d.UUID }

type TitleForm struct {
	Title string `json:"title" jsonschema:"description=What is the title of this document?"`
}

func (d *Document) RegenerateMetadata(ctx context.Context) error {
	var titleForm TitleForm

	history := autoform.MessageSourceFromModelMessages(
		openai.ChatCompletionMessage{
			Role:    "user",
			Content: `Document Content: ` + d.Content,
		},
	)

	if err := autoform.FillForm(ctx, history, &titleForm); err != nil {
		return err
	}

	d.Title = titleForm.Title
	d.Invalidate()

	return d.Update(ctx)
}

var DocumentInterface = psi.DefineNodeInterface[IDocument]()
var DocumentType = psi.DefineNodeType[*Document](psi.WithInterfaceFromNode(DocumentInterface))
var _ IDocument = &Document{}
