package visor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/greenboxal/agibootstrap/pkg/platform/project"
)

type ProjectExplorer struct {
	fyne.CanvasObject
}

func NewProjectExplorer(p project.Project) *ProjectExplorer {
	projectTree := NewPsiTreeWidget(p)
	projectTree.Root = "/"
	projectTree.Refresh()

	projectToolbar := container.NewHBox()

	projectPanel := container.NewVBox(
		projectToolbar,
		projectTree,
	)

	return &ProjectExplorer{
		CanvasObject: container.NewMax(projectPanel),
	}
}
