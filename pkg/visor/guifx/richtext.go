package guifx

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type RichText struct {
	*widget.RichText

	Text obsfx.StringProperty
}

func (rt *RichText) Root() fyne.CanvasObject { return rt.RichText }

func NewRichText() *RichText {
	rt := &RichText{
		RichText: widget.NewRichText(),
	}

	rt.ExtendBaseWidget(rt)

	obsfx.BindFunc(rt.ParseMarkdown, &rt.Text)

	return rt
}
