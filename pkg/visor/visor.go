package visor

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
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

			if id == "" {
				return []string{v.p.UUID()}
			}

			children, err := v.g.GetNodeChildren(id)

			if err != nil {
				return nil
			}

			return lo.Map(children, func(id psi.NodeID, _ int) widget.TreeNodeID {
				return id
			})
		},

		func(id widget.TreeNodeID) bool {
			if id == "" {
				return true
			}

			if v.g == nil {
				return false
			}

			n, err := v.g.GetNode(id)

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

	v.w.SetContent(v.tree)

	return v
}

func (v *Visor) Run() {
	v.w.ShowAndRun()
}

func (v *Visor) Initialize(p project.Project) {
	v.p = p
	v.g = p.Graph().(*indexing.IndexedGraph)

	v.tree.Refresh()

	go func() {
		for range time.Tick(time.Second) {
			v.tree.Refresh()
		}
	}()
}

type ProjectGraph interface {
	GetNode(id psi.NodeID) (psi.Node, error)
	GetNodeChildren(id psi.NodeID) ([]psi.NodeID, error)
}
