package visor

import (
	"context"
	"fmt"
	"path"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/samber/lo"

	"github.com/greenboxal/agibootstrap/pkg/build"
	"github.com/greenboxal/agibootstrap/pkg/build/codegen"
	"github.com/greenboxal/agibootstrap/pkg/build/fiximports"
	"github.com/greenboxal/agibootstrap/pkg/indexing"
	"github.com/greenboxal/agibootstrap/pkg/project"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/tasks"
)

type Visor struct {
	a fyne.App
	w fyne.Window

	tree *widget.Tree
	tabs *container.DocTabs

	p project.Project
	g ProjectGraph
}

func NewVisor(p project.Project) *Visor {
	v := &Visor{}

	v.p = p
	v.a = app.New()
	v.w = v.a.NewWindow("AGIB Visor")

	pathCache := map[string]psi.Node{}

	resolveCached := func(cache map[string]psi.Node, id string) (psi.Node, error) {
		if id == "" {
			panic("empty id")
		}

		if n := cache[id]; n != nil {
			return n, nil
		}

		p := psi.MustParsePath(id)

		n, err := psi.ResolvePath(v.p, p)

		if err != nil {
			return nil, err
		}

		cache[id] = n

		return n, nil
	}

	v.tree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				return []widget.TreeNodeID{"/"}
			}

			resolved, err := resolveCached(pathCache, id)

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

			n, err := resolveCached(pathCache, id)

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

			n, err := resolveCached(pathCache, id)

			if err != nil {
				return
			}

			text := n.String()

			o.(*widget.Label).SetText(text)
		},
	)

	v.tree.Root = "/"
	v.tree.Refresh()

	toolbar := container.NewVBox(
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

	v.tabs = container.NewDocTabs()

	hsplit := container.NewHSplit(
		v.tree,
		v.tabs,
	)

	border := container.NewBorder(
		toolbar,
		nil,
		nil,
		nil,
		hsplit,
	)

	v.w.SetContent(border)

	return v
}

func (v *Visor) Run() {
	v.Initialize()

	v.w.ShowAndRun()
}

func (v *Visor) Initialize() {
	v.g = v.p.Graph().(*indexing.IndexedGraph)

	v.tree.Refresh()
}

type ProjectGraph interface {
	GetNodeByID(id psi.NodeID) (psi.Node, error)
	ResolveNode(id psi.Path) (psi.Node, error)
	GetNodeChildren(id psi.Path) ([]psi.Path, error)
}
