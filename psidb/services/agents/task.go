package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type ITask interface {
	Start(ctx context.Context) error
}

type Task struct {
	psi.NodeBase

	UUID         string           `json:"uuid"`
	Notification psi.Notification `json:"notification"`

	ProgressCurrent int `json:"progress_current"`
	ProgressTotal   int `json:"progress_total"`

	Result *psi.Path `json:"result"`

	Completed    bool   `json:"completed"`
	Error        bool   `json:"error"`
	ErrorMessage string `json:"error_message"`
}

var TaskInterface = psi.DefineNodeInterface[ITask]()
var TaskType = psi.DefineNodeType[*Task](psi.WithInterfaceFromNode(TaskInterface))
var _ ITask = (*Task)(nil)

func (t *Task) PsiNodeName() string { return t.UUID }

func (t *Task) Start(ctx context.Context) error {
	if t.Completed {
		return nil
	}



	return nil
}
