package visor

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/visor/guifx"
)

type SourceFileEditor struct {
	project     project.Project
	elementPath psi.Path
	element     psi.SourceFile

	root fyne.CanvasObject
}

func (t *SourceFileEditor) Project() project.Project { return t.project }
func (t *SourceFileEditor) ElementPath() psi.Path    { return t.elementPath }
func (t *SourceFileEditor) Element() psi.Node        { return t.element }
func (t *SourceFileEditor) Root() fyne.CanvasObject  { return t.root }

func NewSourceFileEditor(p project.Project, elementPath psi.Path, element psi.Node) guifx.Editor {
	tle := &SourceFileEditor{
		project:     p,
		elementPath: elementPath,
		element:     element.(psi.SourceFile),
	}

	textArea := widget.NewRichText()
	textArea.ParseMarkdown(fmt.Sprintf("```\n%s\n```", tle.element.OriginalText()))

	tle.root = container.NewScroll(textArea)

	return tle
}
