package chatui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	"github.com/greenboxal/agibootstrap/pkg/visor/guifx"
)

type ChatHistoryEntry struct {
	root fyne.CanvasObject

	Message obsfx.SimpleProperty[*thoughtdb.Thought]
}

func (ch *ChatHistoryEntry) Root() fyne.CanvasObject { return ch.root }

func NewChatHistoryEntry(t *thoughtdb.Thought) *ChatHistoryEntry {
	ch := &ChatHistoryEntry{}

	ch.Message.SetValue(t)

	textBinding := obsfx.NewObjectBinding[*thoughtdb.Thought, string](
		&ch.Message,
		func(s *thoughtdb.Thought) string {
			if s == nil {
				return ""
			}
			return s.Text
		},
	)

	titleBinding := obsfx.NewObjectBinding[*thoughtdb.Thought, string](
		&ch.Message,
		func(s *thoughtdb.Thought) string {
			if s == nil {
				return ""
			}
			return s.From.Name
		},
	)

	subtitleBinding := obsfx.NewObjectBinding[*thoughtdb.Thought, string](
		&ch.Message,
		func(s *thoughtdb.Thought) string {
			if s == nil {
				return ""
			}
			return s.Pointer.Timestamp.String()
		},
	)

	text := guifx.NewRichText()
	text.Wrapping = fyne.TextWrapWord
	text.Text.Bind(textBinding)

	card := widget.NewCard("", "", text)

	obsfx.BindFunc(card.SetTitle, titleBinding)
	obsfx.BindFunc(card.SetSubTitle, subtitleBinding)

	ch.root = card

	return ch
}
