package visor

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type ChatExplorer struct {
	fyne.CanvasObject
}

func NewChatExplorer(p project.Project, dm *DocumentManager) *ChatExplorer {
	ce := &ChatExplorer{}

	chatLogsToolbar := container.NewHBox()

	chatLogsBinding := binding.NewUntypedList()

	chatLogList := widget.NewListWithData(
		chatLogsBinding,
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(item binding.DataItem, object fyne.CanvasObject) {
			v, err := item.(binding.Untyped).Get()

			if err != nil {
				return
			}

			chatLog, ok := v.(*thoughtstream.ThoughtLog)

			if !ok {
				return
			}

			object.(*widget.Label).SetText(chatLog.Name())
		},
	)

	chatLogList.OnSelected = func(id int) {
		v, err := chatLogsBinding.GetValue(id)

		if err != nil {
			return
		}

		chatLog, ok := v.(*thoughtstream.ThoughtLog)

		if !ok {
			return
		}

		dm.OpenDocument(chatLog.CanonicalPath(), chatLog)
	}

	go func() {
		for range time.Tick(500 * time.Millisecond) {
			c := p.LogManager().PsiNode().Children()
			l := make([]interface{}, len(c))

			for i, v := range c {
				l[i] = v
			}

			chatLogsBinding.Set(l)
		}
	}()

	chatLogsPanel := container.NewBorder(
		chatLogsToolbar,
		nil,
		nil,
		nil,
		chatLogList,
	)

	ce.CanvasObject = chatLogsPanel

	return ce
}

type ThoughtLogEditor struct {
	project     project.Project
	elementPath psi.Path
	element     *thoughtstream.ThoughtLog

	root fyne.CanvasObject
}

func (t *ThoughtLogEditor) Project() project.Project { return t.project }
func (t *ThoughtLogEditor) ElementPath() psi.Path    { return t.elementPath }
func (t *ThoughtLogEditor) Element() psi.Node        { return t.element }
func (t *ThoughtLogEditor) Root() fyne.CanvasObject  { return t.root }

func NewThoughtLogEditor(p project.Project, elementPath psi.Path, element psi.Node) Editor {
	tle := &ThoughtLogEditor{
		project:     p,
		elementPath: elementPath,
		element:     element.(*thoughtstream.ThoughtLog),
	}

	chatMessagesBinding := binding.NewUntypedList()
	chatReplyBinding := binding.NewString()

	listItems := container.NewVBox()
	listContainer := container.NewMax(container.NewVScroll(listItems))

	createItem := func() fyne.CanvasObject {
		rt := widget.NewRichTextFromMarkdown("Message Here")

		rt.Wrapping = fyne.TextWrapWord
		rt.Scroll = container.ScrollNone

		return rt
	}

	updateItem := func(i widget.ListItemID, o fyne.CanvasObject) {
		item, err := chatMessagesBinding.GetItem(i)

		if err != nil {
			fyne.LogError(fmt.Sprintf("Error getting data item %d", i), err)
			return
		}

		v, err := item.(binding.Untyped).Get()

		if err != nil {
			return
		}

		msg, ok := v.(*thoughtstream.Thought)

		if !ok {
			return
		}

		el := o.(*widget.RichText)

		el.ParseMarkdown(fmt.Sprintf("# **[%s]:**\n%s", msg.From.Name, msg.Text))
	}

	updateItems := func() {
		count := chatMessagesBinding.Length()

		for i := 0; i < count; i++ {
			if i >= len(listItems.Objects) {
				listItems.Add(createItem())
			}

			updateItem(i, listItems.Objects[i])
		}

		for i := count; i < len(listItems.Objects); i++ {
			listItems.Remove(listItems.Objects[count])
		}
	}

	chatMessagesBinding.AddListener(binding.NewDataListener(updateItems))

	go func() {
		for range time.Tick(500 * time.Millisecond) {
			c := tle.element.Messages()
			l := make([]interface{}, len(c))

			for i, v := range c {
				n := v
				l[i] = &n
			}

			chatMessagesBinding.Set(l)
		}
	}()

	chatLogDetailsContent := container.NewBorder(
		nil,
		container.NewBorder(
			nil,
			nil,
			nil,
			widget.NewButton("Send", func() {

			}),
			widget.NewEntryWithData(chatReplyBinding),
		),
		nil,
		nil,
		listContainer,
	)

	tle.root = widget.NewCard("Chat Log", "Chat Log details", chatLogDetailsContent)

	return tle
}
