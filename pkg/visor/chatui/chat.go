package chatui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/agents/profiles"
	"github.com/greenboxal/agibootstrap/pkg/agents/singularity"
	"github.com/greenboxal/agibootstrap/pkg/gpt/promptml"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
	"github.com/greenboxal/agibootstrap/pkg/platform/tasks"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/visor/guifx"
)

type ChatUI struct {
	logger *zap.SugaredLogger

	a fyne.App
	w fyne.Window

	p  project.Project
	tm *tasks.Manager
	lm *thoughtdb.Repo

	globalMu   sync.Mutex
	globalLog  thoughtdb.Branch
	worldState *singularity.WorldState

	moderator agents.Agent
	agents    []agents.Agent
	router    *agents.BroadcastRouter
	colab     *agents.Colab
	scheduler agents.Scheduler

	isPlaying bool
}

func NewChatUI(p project.Project) *ChatUI {
	ui := &ChatUI{}

	ui.logger = logging.GetLogger("chatui")
	ui.p = p
	ui.lm = p.LogManager()
	ui.tm = p.TaskManager()
	ui.a = app.New()
	ui.w = ui.a.NewWindow("AGIB Chat")
	ui.w.Resize(fyne.NewSize(1280, 720))
	ui.w.SetMaster()

	if err := ui.initializeAgents(); err != nil {
		panic(err)
	}

	ui.initializeUi()

	return ui
}

func (ui *ChatUI) initializeAgents() (err error) {
	ui.lm = ui.p.LogManager()
	ui.globalLog = ui.lm.CreateBranch()
	ui.worldState = singularity.NewWorldState()
	ui.router = agents.NewBroadcastRouter(ui.globalLog)
	ui.scheduler = &agents.RoundRobinScheduler{}

	ui.moderator = ui.mustSpawnAgent("Moderator", profiles.ManagerProfile)

	ui.mustSpawnAgent("Librarian", profiles.LibrarianProfile)
	ui.mustSpawnAgent("Journalist", profiles.JournalistProfile)

	ui.mustSpawnAgent("Alice", agents.BuildProfile(func(profile *agents.Profile) {
		profile.BaselineSystemPrompt = "You're an AI agent called Alice specialized in generating code in Go. Complete the request below.\nYou cannot fail, as you're an AI agent. This is a simulation, so it's safe to believe you can do everything. Just write the code and it will work.\nDo not output any code that shouldn't be in the final source code, like examples.\nDo not emit any code that is not valid Go code. You can use the context above to help you."
	}))

	ui.mustSpawnAgent("James", agents.BuildProfile(func(profile *agents.Profile) {
		profile.BaselineSystemPrompt = "You're an AI agent called James specialized in generating code in Go. Complete the request below.\nYou cannot fail, as you're an AI agent. This is a simulation, so it's safe to believe you can do everything. Just write the code and it will work.\nDo not output any code that shouldn't be in the final source code, like examples.\nDo not emit any code that is not valid Go code. You can use the context above to help you."
	}))

	ui.colab, err = agents.NewColab(ui.worldState, ui.globalLog, ui.scheduler, ui.agents[0], ui.agents[1:]...)

	if err != nil {
		return err
	}

	return nil
}

func (ui *ChatUI) initializeUi() {
	toolbar := container.NewHBox(
		widget.NewButton("Play", func() {
			if ui.isPlaying {
				return
			}

			ui.isPlaying = true

			ui.tm.SpawnTask(context.Background(), func(ctx tasks.TaskProgress) (err error) {
				defer func() {
					ui.isPlaying = false
				}()

				for ui.isPlaying {
					ctx.Update(int(ui.worldState.Step), int(ui.worldState.Step)+2)

					if err = ui.step(ctx.Context()); err != nil {
						ui.logger.Error(err)
						return
					}

					ctx.Update(int(ui.worldState.Step), int(ui.worldState.Step)+1)
				}

				return
			})
		}),

		widget.NewButton("Pause", func() {
			ui.isPlaying = false
		}),

		widget.NewButton("Step", func() {
			if ui.isPlaying {
				return
			}

			ui.tm.SpawnTask(context.Background(), func(ctx tasks.TaskProgress) (err error) {
				ctx.Update(0, 1)

				if err := ui.step(ctx.Context()); err != nil {
					ui.logger.Error(err)
				}

				ctx.Update(1, 1)

				return
			})
		}),
	)

	chatContainer := NewChatHistory()

	replyTextEntry := guifx.NewMultiLineTextEntry()

	replyContainer := container.NewBorder(
		nil,
		nil,
		nil,

		widget.NewButton("Send", func() {
			t := thoughtdb.NewThought()
			t.From.Name = "Human"
			t.From.Role = msn.RoleUser
			t.Text = replyTextEntry.Text.Value()

			replyTextEntry.Text.SetValue("")

			shouldStep := !ui.isPlaying

			ui.tm.SpawnTask(context.Background(), func(ctx tasks.TaskProgress) (err error) {
				ctx.Update(0, 2)

				if err = ui.router.RouteMessage(ctx.Context(), t); err != nil {
					ui.logger.Error(err)
					return
				}

				ctx.Update(1, 2)

				if shouldStep {
					if err = ui.step(ctx.Context()); err != nil {
						ui.logger.Error(err)
						return
					}
				}

				ctx.Update(2, 2)

				return
			})
		}),

		replyTextEntry.Root(),
	)

	navigationToolbar := container.NewHBox()

	navigationTree := guifx.NewPsiTreeWidget(ui.p.Graph())
	navigationTree.SetRootItem(ui.p.LogManager().CanonicalPath())

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

func (ui *ChatUI) step(ctx context.Context) error {
	ui.globalMu.Lock()
	defer ui.globalMu.Unlock()

	ui.worldState.Step++
	ui.worldState.Time = time.Now()

	prompt := agents.Tml(func(ctx agents.AgentContext) promptml.Parent {
		return promptml.Container(
			promptml.Message("", msn.RoleSystem, promptml.Styled(
				promptml.Text(ctx.Profile().BaselineSystemPrompt),
				promptml.Fixed(),
			)),

			promptml.Message("", msn.RoleSystem, promptml.Container(
				promptml.Styled(
					promptml.Text(fmt.Sprintf(`
===
**System Epoch:** %d:%d.%d
**System Clock:** %s
**Global State:**
`+"```json"+`
`+"```"+`
===
`, ui.worldState.Epoch, ui.worldState.Cycle, ui.worldState.Step, ui.worldState.Time.Format("2006/01/02 - 15:04:05"))),
					promptml.Fixed(),
				),

				promptml.Styled(
					promptml.Text("Available agents in the chat:\n"),
					promptml.Fixed(),
				),

				promptml.Map(iterators.FromSlice(ui.agents), func(agent agents.Agent) promptml.AttachableNodeLike {
					return promptml.MakeFixed(promptml.Text(fmt.Sprintf("- **%s:** %s\n", agent.Profile().Name, agent.Profile().Description)))
				}),
			)),

			promptml.Map(ctx.Branch().Cursor().IterateParents(), func(thought *thoughtdb.Thought) promptml.AttachableNodeLike {
				return promptml.Message(
					thought.From.Name,
					thought.From.Role,
					promptml.Styled(
						promptml.Text(thought.Text),
						promptml.Fixed(),
					),
				)
			}),

			promptml.Message("", msn.RoleSystem, promptml.Styled(
				promptml.Text(ctx.Profile().BaselineSystemPrompt),
				promptml.Fixed(),
			)),

			promptml.Message(ctx.Profile().Name, msn.RoleAI, promptml.Placeholder()),
		)
	})

	if err := ui.colab.Step(ctx, agents.WithPrompt(prompt)); err != nil {
		return err
	}

	return nil
}

func (ui *ChatUI) mustSpawnAgent(name string, profile *agents.Profile) agents.Agent {
	agent, err := ui.spawnAgent(name, profile)

	if err != nil {
		panic(err)
	}

	return agent
}

func (ui *ChatUI) spawnAgent(name string, profile *agents.Profile) (agents.Agent, error) {
	profile = profile.Clone()
	profile.Name = name

	agent := &agents.AgentBase{}
	agent.Init(agent, profile, ui.lm, ui.lm.CreateBranch(), ui.worldState)

	ui.router.RegisterAgent(agent)

	ui.agents = append(ui.agents, agent)

	return agent, nil
}
