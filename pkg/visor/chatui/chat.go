package chatui

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/agents/profiles"
	"github.com/greenboxal/agibootstrap/pkg/agents/singularity"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/visor/guifx"
)

type ChatUI struct {
	a fyne.App
	w fyne.Window

	p project.Project

	globalLog  thoughtdb.Branch
	rootBranch thoughtdb.Branch
	worldState *singularity.WorldState

	agent  *agents.AgentBase
	router *agents.BroadcastRouter
}

func NewChatUI(p project.Project) *ChatUI {
	ui := &ChatUI{}

	ui.p = p
	ui.a = app.New()
	ui.w = ui.a.NewWindow("AGIB Chat")
	ui.w.Resize(fyne.NewSize(1280, 720))
	ui.w.SetMaster()

	ui.initializeAgents()
	ui.initializeUi()

	return ui
}

func (ui *ChatUI) initializeAgents() {
	lm := ui.p.LogManager()

	globalLog := lm.CreateBranch()
	branch := lm.CreateBranch()

	ui.rootBranch = branch
	ui.globalLog = globalLog
	ui.worldState = singularity.NewWorldState()
	ui.router = agents.NewBroadcastRouter(ui.globalLog)

	ui.agent = &agents.AgentBase{}
	ui.agent.Init(ui.agent, profiles.PairProfile, nil, ui.rootBranch, ui.worldState)

	ui.router.RegisterAgent(ui.agent)
}

func (ui *ChatUI) initializeUi() {
	toolbar := container.NewHBox()

	chatContainer := NewChatHistory()

	replyTextEntry := guifx.NewMultiLineTextEntry()

	replyContainer := container.NewBorder(
		nil,
		nil,
		nil,

		widget.NewButton("Send", func() {
			t := thoughtdb.NewThought()
			t.From.Name = "You"
			t.Pointer.Timestamp = time.Now()
			t.Text = replyTextEntry.Text.Value()

			replyTextEntry.Text.SetValue("")

			ctx := context.Background()

			if err := ui.router.RouteMessage(ctx, t); err != nil {
				panic(err)
			}

			if err := ui.router.RouteIncomingMessages(ctx); err != nil {
				panic(err)
			}

			if err := ui.agent.Step(ctx); err != nil {
				panic(err)
			}
		}),

		replyTextEntry.Root(),
	)

	navigationToolbar := container.NewHBox()

	navigationTree := guifx.NewPsiTreeWidget(ui.p)
	navigationTree.Root = ui.p.LogManager().CanonicalPath().String()

	navigationContainer := container.NewBorder(
		navigationToolbar,
		nil,
		nil,
		nil,
		navigationTree,
	)

	root := container.NewBorder(
		toolbar,
		replyContainer,
		navigationContainer,
		nil,
		chatContainer.Root(),
	)

	ui.w.SetContent(root)

	collectionsfx.BindList(&chatContainer.Entries, ui.globalLog.ChildrenList(), func(v psi.Node) *thoughtdb.Thought {
		t, ok := v.(*thoughtdb.Thought)

		if !ok {
			return nil
		}

		return t
	})
}

func (ui *ChatUI) Run() {
	ui.w.ShowAndRun()
}
