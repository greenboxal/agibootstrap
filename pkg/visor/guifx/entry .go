package guifx

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type TextEntry struct {
	Entry *widget.Entry

	Text obsfx.StringProperty
}

func (e *TextEntry) Root() fyne.CanvasObject { return e.Entry }

func wrapTextEntry(entry *widget.Entry) *TextEntry {
	e := &TextEntry{
		Entry: entry,
	}

	e.Entry.OnChanged = func(text string) {
		e.Text.SetValue(text)
	}

	obsfx.ObserveChange(&e.Text, func(old, new string) {
		e.Entry.SetText(new)
	})

	return e
}

func NewTextEntry() *TextEntry {
	return wrapTextEntry(widget.NewEntry())
}

func NewMultiLineTextEntry() *TextEntry {
	return wrapTextEntry(widget.NewMultiLineEntry())
}
