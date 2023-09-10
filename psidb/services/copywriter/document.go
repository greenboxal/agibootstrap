package copywriter

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type AppendDocumentTextRequest struct {
	Heading      string `json:"heading"`
	HeadingLevel int    `json:"heading_level"`
	Markdown     string `json:"markdown"`
}

type SelectHeadingRequest struct {
	Heading      string `json:"heading"`
	HeadingLevel int    `json:"heading_level"`
}

type IDocumentEditor interface {
	GetDocumentContent(ctx context.Context) (string, error)
	GetCurrentHeading(ctx context.Context) (DocumentHeading, error)

	SelectRootHeading(ctx context.Context) (DocumentHeading, error)
	SelectParentHeading(ctx context.Context) (DocumentHeading, error)
	SelectHeading(ctx context.Context, req SelectHeadingRequest) (DocumentHeading, error)

	AppendDocumentText(ctx context.Context, req AppendDocumentTextRequest) (DocumentHeading, error)
}

var DocumentEditorInterface = psi.DefineNodeInterface[IDocumentEditor]()

type DocumentHeading struct {
	Index         int    `json:"index"`
	Order         int    `json:"order"`
	ParentSection int    `json:"parent_section"`
	Heading       string `json:"heading"`
	HeadingLevel  int    `json:"heading_level"`
	Content       string `json:"content"`
}

type DocumentEditor struct {
	psi.NodeBase

	Name           string            `json:"name"`
	Headings       []DocumentHeading `json:"sections"`
	CurrentHeading int               `json:"last_section,omitempty"`
}

var _ IDocumentEditor = (*DocumentEditor)(nil)

func (d *DocumentEditor) PsiNodeName() string { return d.Name }

func (d *DocumentEditor) GetCurrentHeading(ctx context.Context) (DocumentHeading, error) {
	return d.getHeadingByIndexOrNil(d.CurrentHeading), nil
}

func (d *DocumentEditor) SelectRootHeading(ctx context.Context) (DocumentHeading, error) {
	return d.getHeadingByIndexOrNil(0), nil
}

func (d *DocumentEditor) SelectParentHeading(ctx context.Context) (DocumentHeading, error) {
	current := d.getHeadingByIndexOrNil(d.CurrentHeading)

	return d.getHeadingByIndexOrNil(current.ParentSection), nil
}

// Write the design document for the Go programming language, step by step, using IDocumentEditor_QZQZ_AppendDocumentText

func (d *DocumentEditor) AppendDocumentText(ctx context.Context, req AppendDocumentTextRequest) (DocumentHeading, error) {
	section := DocumentHeading{
		Heading:      req.Heading,
		HeadingLevel: req.HeadingLevel,
		Content:      req.Markdown,
	}

	return d.appendHeading(ctx, section)
}

func (d *DocumentEditor) SelectHeading(ctx context.Context, req SelectHeadingRequest) (DocumentHeading, error) {
	for i, heading := range d.Headings {
		if heading.Heading == req.Heading && heading.HeadingLevel == req.HeadingLevel {
			d.CurrentHeading = i
			return heading, nil
		}
	}

	heading := DocumentHeading{
		Heading:      req.Heading,
		HeadingLevel: req.HeadingLevel,
	}

	return d.appendHeading(ctx, heading)
}

func (d *DocumentEditor) appendHeading(ctx context.Context, heading DocumentHeading) (DocumentHeading, error) {
	if len(d.Headings) == 0 {
		heading.ParentSection = -1
	} else {
		current := d.getHeadingByIndexOrNil(d.CurrentHeading)

		if heading.HeadingLevel > current.HeadingLevel {
			heading.ParentSection = d.CurrentHeading
			heading.Order = current.Order + 1
		} else {
			heading.ParentSection = current.ParentSection
			heading.Order = current.Order + 1
		}

		heading.ParentSection = d.CurrentHeading
	}

	heading.Index = len(d.Headings)
	d.Headings = append(d.Headings, heading)
	d.CurrentHeading = heading.Index

	d.Invalidate()

	if err := d.Update(ctx); err != nil {
		return DocumentHeading{}, err
	}

	return heading, nil
}

func (d *DocumentEditor) getHeadingByIndexOrNil(index int) DocumentHeading {
	if index == -1 {
		return DocumentHeading{}
	}

	if d.CurrentHeading >= len(d.Headings) {
		return DocumentHeading{}
	}

	return d.Headings[index]
}

func (d *DocumentEditor) compareHeadingInTree(a, b DocumentHeading) bool {
	if a.ParentSection == b.ParentSection {
		return a.Order < b.Order
	} else {
		ap := a
		bp := b

		if ap.ParentSection != -1 && ap.ParentSection != ap.Index {
			ap = d.getHeadingByIndexOrNil(ap.ParentSection)
		}

		if bp.ParentSection != -1 && bp.ParentSection != bp.Index {
			bp = d.getHeadingByIndexOrNil(bp.ParentSection)
		}

		return d.compareHeadingInTree(ap, bp)
	}
}

func (d *DocumentEditor) GetDocumentContent(ctx context.Context) (string, error) {
	buffer := &bytes.Buffer{}

	headings := slices.Clone(d.Headings)

	slices.SortFunc(headings, d.compareHeadingInTree)

	for _, heading := range headings {
		buffer.WriteString(fmt.Sprintf("%s %s\n\n", strings.Repeat("#", heading.HeadingLevel+1), heading.Heading))
		buffer.WriteString(heading.Content)
		buffer.WriteString("\n")
	}

	return buffer.String(), nil
}

var DocumentEditorType = psi.DefineNodeType[*DocumentEditor](psi.WithInterfaceFromNode(DocumentEditorInterface))
