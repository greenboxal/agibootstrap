package visor

import (
	"context"
	"fmt"
	"path"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/samber/lo"

	"github.com/greenboxal/agibootstrap/pkg/build"
	"github.com/greenboxal/agibootstrap/pkg/build/codegen"
	"github.com/greenboxal/agibootstrap/pkg/build/fiximports"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/tasks"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Visor struct {
	a fyne.App
	w fyne.Window

	tree *PsiTreeWidget
	tabs *container.DocTabs

	p project.Project
}

func NewVisor(p project.Project) *Visor {
	v := &Visor{}

	v.p = p
	v.a = app.New()
	v.w = v.a.NewWindow("AGIB Visor")
	v.w.Resize(fyne.NewSize(1280, 720))

	v.tree = NewPsiTreeWidget(p)
	v.tree.Root = "/"
	v.tree.Refresh()

	mainToolbar := container.NewHBox(
		widget.NewButton("Refresh", func() {
			v.tree.Refresh()
		}),

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
	)

	tasksPanelToolbar := container.NewHBox()

	selectedTaskId := binding.NewString()
	selectedTaskName := binding.NewString()
	selectedTaskDesc := binding.NewString()
	selectedTaskProgress := binding.NewFloat()

	taskTree := NewPsiTreeWidget(p.TaskManager().PsiNode())

	updateSelectedTask := func() {
		v, err := taskTree.SelectedItem.Get()

		if err != nil {
			return
		}

		task, ok := v.(tasks.Task)

		if !ok {
			return
		}

		selectedTaskId.Set(task.UUID())
		selectedTaskName.Set(task.Name())
		selectedTaskDesc.Set(task.Description())
		selectedTaskProgress.Set(task.Progress())
	}

	go func() {
		for range time.Tick(500 * time.Millisecond) {
			updateSelectedTask()
		}
	}()

	taskTree.SelectedItem.AddListener(binding.NewDataListener(func() {
		updateSelectedTask()
	}))

	taskDetailsContent := container.NewVBox(
		widget.NewLabelWithData(selectedTaskId),
		widget.NewLabelWithData(selectedTaskName),
		widget.NewLabelWithData(selectedTaskDesc),
		widget.NewProgressBarWithData(selectedTaskProgress),
	)

	taskDetails := widget.NewCard("Task", "Task details", taskDetailsContent)

	tasksPanel := container.NewBorder(
		tasksPanelToolbar,
		nil,
		nil,
		nil,
		container.NewHSplit(
			taskTree,
			taskDetails,
		),
	)

	v.tabs = container.NewDocTabs()

	leftDrawer := container.NewAppTabs(
		container.NewTabItem("Project", v.tree),
	)

	rightDrawer := container.NewAppTabs()

	bottomDrawer := container.NewAppTabs(
		container.NewTabItem("Tasks", tasksPanel),
	)

	hsplit := container.NewHSplit(
		leftDrawer,
		container.NewHSplit(
			v.tabs,
			rightDrawer,
		),
	)

	vsplit := container.NewVSplit(
		hsplit,
		bottomDrawer,
	)

	border := container.NewBorder(
		mainToolbar,
		nil,
		nil,
		nil,
		vsplit,
	)

	v.w.SetContent(border)

	return v
}

func (v *Visor) Run() {
	v.Initialize()

	v.w.ShowAndRun()
}

func (v *Visor) Initialize() {
	v.tree.Refresh()
}

type PsiTreeWidget struct {
	*widget.Tree

	root  psi.Node
	cache map[string]psi.Node

	SelectedItem binding.Untyped
}

func (w *PsiTreeWidget) Node(id widget.TreeNodeID) psi.Node {
	n, err := w.resolveCached(id)

	if err != nil {
		return nil
	}

	return n
}

func (w *PsiTreeWidget) resolveCached(id string) (psi.Node, error) {
	if id == "" {
		panic("empty id")
	}

	if n := w.cache[id]; n != nil {
		return n, nil
	}

	p := psi.MustParsePath(id)

	n, err := psi.ResolvePath(w.root, p)

	if err != nil {
		return nil, err
	}

	w.cache[id] = n

	return n, nil
}

func NewPsiTreeWidget(root psi.Node) *PsiTreeWidget {
	ptw := &PsiTreeWidget{
		root:  root,
		cache: map[string]psi.Node{},

		SelectedItem: binding.NewUntyped(),
	}

	ptw.Tree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				return []widget.TreeNodeID{"/"}
			}

			resolved, err := ptw.resolveCached(id)

			if err != nil {
				return nil
			}

			return lo.Map(resolved.Children(), func(child psi.Node, _ int) widget.TreeNodeID {
				p := child.CanonicalPath()

				if p.IsEmpty() {
					panic("empty path")
				}

				ps := p.String()

				if ps == "" {
					panic("empty path")
				}

				return ps
			})
		},

		func(id widget.TreeNodeID) bool {
			if id == "" {
				return true
			}

			n, err := ptw.resolveCached(id)

			if err != nil {
				return false
			}

			return n.IsContainer()
		},

		func(branch bool) fyne.CanvasObject {
			if branch {
				return widget.NewLabel("Branch template")
			}

			return widget.NewLabel("Leaf template")
		},

		func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			if id == "" {
				return
			}

			n, err := ptw.resolveCached(id)

			if err != nil {
				return
			}

			text := n.String()

			o.(*widget.Label).SetText(text)
		},
	)

	ptw.Tree.OnSelected = func(id widget.TreeNodeID) {
		if err := ptw.SelectedItem.Set(ptw.Node(id)); err != nil {
			panic(err)
		}
	}

	return ptw
}
