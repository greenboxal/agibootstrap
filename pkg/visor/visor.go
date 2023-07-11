package visor

import (
	"context"
	"fmt"
	"path"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"

	"github.com/greenboxal/agibootstrap/pkg/build"
	"github.com/greenboxal/agibootstrap/pkg/build/codegen"
	"github.com/greenboxal/agibootstrap/pkg/build/fiximports"
	"github.com/greenboxal/agibootstrap/pkg/gpt/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt/agents/profiles"
	singularity2 "github.com/greenboxal/agibootstrap/pkg/gpt/agents/singularity"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/tasks"
)

type Visor struct {
	a fyne.App
	w fyne.Window

	dm *DocumentManager

	p project.Project
}

func NewVisor(p project.Project) *Visor {
	v := &Visor{}

	v.p = p
	v.a = app.New()
	v.w = v.a.NewWindow("AGIB Visor")
	v.w.Resize(fyne.NewSize(1280, 720))
	v.w.SetMaster()

	v.dm = NewDocumentManager(v.p)

	mainToolbar := container.NewHBox(
		widget.NewButton("Build", func() {
			v.p.TaskManager().SpawnTask(context.Background(), func(progress tasks.TaskProgress) error {
				builder := build.NewBuilder(v.p, build.Configuration{
					OutputDirectory: v.p.RootPath(),
					BuildDirectory:  path.Join(v.p.RootPath(), ".build"),

					BuildSteps: []build.Step{
						&codegen.BuildStep{},
						&fiximports.BuildStep{},
					},
				})

				_, err := builder.Build(progress.Context())

				if err != nil {
					fmt.Printf("error: %s\n", err)
				}

				return nil
			})
		}),

		widget.NewButton("Boot", func() {
			v.p.TaskManager().SpawnTask(context.Background(), func(tctx tasks.TaskProgress) error {
				s, err := singularity2.NewSingularity(p.LogManager())

				if err != nil {
					return err
				}

				s.SetParent(v.p)

				request := thoughtdb.NewThought()
				request.Pointer.Timestamp = time.Now()
				request.From.Name = "Human"
				request.From.Role = msn.RoleUser
				request.Text = `
Create a Pytorch model based on the human brain cytoarchitecture.
`

				s.Router().ReceiveIncomingMessage(tctx.Context(), request)

				st := s.WorldState().(*singularity2.WorldState)

				for {
					fmt.Printf("Singularity: Step (epoch = %d, cycle = %d, step = %d)", st.Epoch, st.Cycle, st.Step)
					_, err := s.Step(tctx.Context())

					if err != nil {
						return err
					}

					progress := agents.GetState(st, profiles.CtxGoalStatus)

					if progress.Completed {
						break
					}
				}

				return nil
			})
		}),
	)

	documentArea := NewDocumentArea(v.dm)

	projectExplorer := NewProjectExplorer(p, v.dm)
	chatExplorer := NewChatExplorer(p, v.dm)
	tasksToolWindow := NewTasksToolWindow(p)
	propertyInspector := NewPropertyInspector()

	leftDrawer := container.NewAppTabs(
		container.NewTabItem("Project", projectExplorer.CanvasObject),
		container.NewTabItem("Chats", chatExplorer.CanvasObject),
	)

	rightDrawer := container.NewAppTabs(
		container.NewTabItem("Properties", propertyInspector.CanvasObject),
	)

	bottomDrawer := container.NewAppTabs(
		container.NewTabItem("Tasks", tasksToolWindow.CanvasObject),
	)

	bottomDrawer.SetTabLocation(container.TabLocationBottom)

	splitGrid := container.NewHSplit(
		leftDrawer,
		documentArea,
	)

	splitGrid.Offset = 0.2

	container.NewHSplit(
		rightDrawer,
		rightDrawer,
	)

	border := container.NewBorder(
		mainToolbar,
		container.NewMax(bottomDrawer),
		nil,
		nil,
		container.NewMax(splitGrid),
	)

	v.w.SetContent(border)

	return v
}

func (v *Visor) Run() {
	v.w.ShowAndRun()
}
