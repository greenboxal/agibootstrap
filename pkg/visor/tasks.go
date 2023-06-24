package visor

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/tasks"
)

type TasksToolWindow struct {
	fyne.CanvasObject
}

func NewTasksToolWindow(p project.Project) *TasksToolWindow {
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

	return &TasksToolWindow{
		CanvasObject: tasksPanel,
	}
}
