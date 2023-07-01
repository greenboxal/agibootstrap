package visor

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	obsfx "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	collectionsfx2 "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/visor/guifx"
)

type ChatExplorer struct {
	fyne.CanvasObject
}

func NewChatExplorer(p project.Project, dm *DocumentManager) *ChatExplorer {
	ce := &ChatExplorer{}

	chatLogsToolbar := container.NewHBox()

	chatLogTree := guifx.NewPsiTreeWidget(p)
	chatLogTree.SetRootItem(p.LogManager().CanonicalPath())

	chatLogTree.OnNodeSelected = func(n psi.Node) {
		chatLog, ok := n.(thoughtdb.Branch)

		if !ok {
			return
		}

		dm.OpenDocument(chatLog.CanonicalPath(), chatLog)
	}

	chatLogsPanel := container.NewBorder(
		chatLogsToolbar,
		nil,
		nil,
		nil,
		chatLogTree,
	)

	ce.CanvasObject = chatLogsPanel

	return ce
}

type ThoughtLogEditor struct {
	project     project.Project
	elementPath psi.Path
	element     thoughtdb.Branch

	root fyne.CanvasObject
}

func (t *ThoughtLogEditor) Project() project.Project { return t.project }
func (t *ThoughtLogEditor) ElementPath() psi.Path    { return t.elementPath }
func (t *ThoughtLogEditor) Element() psi.Node        { return t.element }
func (t *ThoughtLogEditor) Root() fyne.CanvasObject  { return t.root }

func NewThoughtLogEditor(p project.Project, elementPath psi.Path, element psi.Node) guifx.Editor {
	tle := &ThoughtLogEditor{
		project:     p,
		elementPath: elementPath,
		element:     element.(thoughtdb.Branch),
	}

	chatReplyBinding := binding.NewString()

	listItemParent := container.NewVBox()
	listContainer := container.NewMax(container.NewVScroll(listItemParent))

	thoughtList := collectionsfx2.MutableSlice[*thoughtdb.Thought]{}
	listItems := collectionsfx2.MutableSlice[*ThoughtView]{}

	collectionsfx2.ObserveList(&listItems, func(ev collectionsfx2.ListChangeEvent[*ThoughtView]) {
		for ev.Next() {
			if ev.WasRemoved() {
				for _, removed := range ev.RemovedSlice() {
					listItemParent.Remove(removed.View)
				}
			} else {
				for _, added := range ev.AddedSlice() {
					listItemParent.Add(added.View)
				}
			}
		}
	})

	collectionsfx2.BindList(&listItems, &thoughtList, func(v *thoughtdb.Thought) *ThoughtView {
		tv := NewThoughtView()

		tv.Thought.SetValue(v)

		return tv
	})

	updateAllItems := func() {
		children := iterators.ToSlice(iterators.FilterIsInstance[psi.Node, *thoughtdb.Thought](tle.element.ChildrenIterator()))

		thoughtList.ReplaceAll(children...)
	}

	obsfx.ObserveInvalidation(element.ChildrenList(), updateAllItems)

	updateAllItems()

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

type ThoughtView struct {
	View fyne.CanvasObject

	Thought      obsfx.SimpleProperty[*thoughtdb.Thought]
	TextProperty obsfx.StringProperty

	rt *widget.RichText
}

func NewThoughtView() *ThoughtView {
	tv := &ThoughtView{}

	tv.rt = widget.NewRichText()
	tv.rt.Wrapping = fyne.TextWrapWord
	tv.rt.Scroll = container.ScrollNone

	tv.View = tv.rt

	obsfx.BindFunc(func(v string) {
		msg := tv.Thought.Value()

		if msg == nil {
			return
		}

		tv.rt.ParseMarkdown(fmt.Sprintf("# **[%s]:**\n%s", msg.From.Name, v))
	}, &tv.TextProperty)

	obsfx.ObserveChange(&tv.Thought, func(old, new *thoughtdb.Thought) {
		if old != nil {
			old.ChildrenList().RemoveListener(tv)
		}

		if new != nil {
			new.ChildrenList().AddListener(tv)
		}

		tv.OnInvalidated(nil)
	})

	tv.OnInvalidated(nil)

	return tv
}

func (tv *ThoughtView) OnInvalidated(o obsfx.Observable) {
	t := tv.Thought.Value()

	if t != nil {
		tv.TextProperty.SetValue(t.Text)
	} else {
		tv.TextProperty.SetValue("")
	}
}
