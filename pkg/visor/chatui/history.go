package chatui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
)

type ChatHistory struct {
	root fyne.CanvasObject

	Entries collectionsfx.MutableSlice[*thoughtdb.Thought]
}

func (ch *ChatHistory) Root() fyne.CanvasObject { return ch.root }

func NewChatHistory() *ChatHistory {
	ch := &ChatHistory{}

	entriesContainer := container.NewVBox()
	entriesScroll := container.NewVScroll(entriesContainer)

	collectionsfx.ObserveList(&ch.Entries, func(ev collectionsfx.ListChangeEvent[*thoughtdb.Thought]) {
		bottomOffsetY := -1 * (entriesScroll.Content.MinSize().Height - entriesScroll.Size().Height)
		wasAtBottom := entriesScroll.Offset.Y == bottomOffsetY

		for ev.Next() {
			if ev.WasAdded() {
				for i, item := range ev.AddedSlice() {
					if item == nil {
						entriesContainer.Objects = slices.Insert(entriesContainer.Objects, ev.From()+i, fyne.CanvasObject(container.NewVBox()))

						return
					}

					entry := NewChatHistoryEntry(item)
					entry.Message.SetValue(item)

					entriesContainer.Objects = slices.Insert(entriesContainer.Objects, ev.From()+i, entry.Root())
				}
			} else if ev.WasRemoved() {
				for i := 0; i < ev.RemovedCount(); i++ {
					obj := entriesContainer.Objects[ev.From()]

					entriesContainer.Remove(obj)
				}
			}
		}

		entriesContainer.Refresh()

		if wasAtBottom {
			entriesScroll.ScrollToBottom()
		}
	})

	ch.root = entriesScroll

	return ch
}
