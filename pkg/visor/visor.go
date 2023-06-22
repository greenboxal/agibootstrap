package visor

import (
	"fmt"
	"path"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/samber/lo"

	"github.com/greenboxal/agibootstrap/pkg/indexing"
	"github.com/greenboxal/agibootstrap/pkg/project"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

var globalVisor *Visor

type Visor struct {
	a    fyne.App
	w    fyne.Window
	p    project.Project
	g    ProjectGraph
	tree *widget.Tree
}

func NewVisor() *Visor {
	if globalVisor == nil {
		globalVisor = newVisor()
	}

	return globalVisor
}

func newVisor() *Visor {
	v := &Visor{}

	v.a = app.New()
	v.w = v.a.NewWindow("AGIB Visor")

	v.tree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if v.g == nil || v.p == nil {
				return nil
			}

			rootPath := fmt.Sprintf("Project:%s@0", v.p.UUID())

			if id == "" {
				return []string{rootPath}
			}

			p := psi.ParsePath(id)

			children, err := v.g.GetNodeChildren(p)

			if err != nil {
				return nil
			}

			return lo.Map(children, func(id psi.Path, _ int) widget.TreeNodeID {
				return path.Join(rootPath, id.String())
			})
		},

		func(id widget.TreeNodeID) bool {
			if id == "" {
				return true
			}

			if v.g == nil {
				return false
			}

			p := psi.ParsePath(id)
			n, err := v.g.ResolveNode(p)

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
			text := id

			if branch {
				text += " (branch)"
			}

			o.(*widget.Label).SetText(text)
		},
	)

	vbox := container.NewVBox(
		widget.NewButton("Refresh", func() {
			v.tree.Refresh()
		}),

		v.tree,
	)

	v.w.SetContent(vbox)

	return v
}

func (v *Visor) Run() {
	v.w.ShowAndRun()
}

func (v *Visor) Initialize(p project.Project) {
	v.p = p
	v.g = p.Graph().(*indexing.IndexedGraph)

	v.tree.Refresh()
}

type ProjectGraph interface {
	GetNodeByID(id psi.NodeID) (psi.Node, error)
	ResolveNode(id psi.Path) (psi.Node, error)
	GetNodeChildren(id psi.Path) ([]psi.Path, error)
}
