package visor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/greenboxal/agibootstrap/pkg/platform/project"
)

type ProjectExplorer struct {
	fyne.CanvasObject
}

func NewProjectExplorer(p project.Project, dm *DocumentManager) *ProjectExplorer {
	projectTree := NewPsiTreeWidget(p)

	projectToolbar := container.NewHBox(
		widget.NewButton("Refresh", func() {
			projectTree.pathCache = map[string]*psiTreeNodeState{}
			projectTree.Refresh()
		}),
	)

	projectPanel := container.NewBorder(
		projectToolbar,
		nil,
		nil,
		nil,
		projectTree,
	)

	return &ProjectExplorer{
		CanvasObject: projectPanel,
	}
}
