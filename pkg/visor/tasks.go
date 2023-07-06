package visor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/tasks"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/visor/guifx"
)

type TasksToolWindow struct {
	fyne.CanvasObject
}

func NewTasksToolWindow(p project.Project) *TasksToolWindow {
	var selectedTask tasks.Task

	tasksPanelToolbar := container.NewHBox()

	selectedTaskId := binding.NewString()
	selectedTaskName := binding.NewString()
	selectedTaskDesc := binding.NewString()
	selectedTaskProgress := binding.NewFloat()

	taskTree := guifx.NewPsiTreeWidget(p.Graph())
	taskTree.SetRootItem(p.TaskManager().CanonicalPath())

	updateSelectedTask := func() {
		if selectedTask != nil {
			selectedTaskId.Set(selectedTask.CanonicalPath().String())
			selectedTaskName.Set(selectedTask.Name())
			selectedTaskDesc.Set(selectedTask.Description())
			selectedTaskProgress.Set(selectedTask.Progress())
		}
	}

	taskTree.OnNodeSelected = func(n psi.Node) {
		task, ok := n.(tasks.Task)

		if !ok {
			return
		}

		if selectedTask == task {
			return
		}

		if selectedTask != nil {
			//selectedTask.ChildrenList(invalidationListener)
		}

		selectedTask = task

		if selectedTask != nil {
			//selectedTask.AddInvalidationListener(invalidationListener)
		}

		updateSelectedTask()
	}

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
			container.NewScroll(taskTree),
			taskDetails,
		),
	)

	return &TasksToolWindow{
		CanvasObject: tasksPanel,
	}
}
