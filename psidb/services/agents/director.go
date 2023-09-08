package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type CheckListItem struct {
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

type CheckList struct {
	psi.NodeBase

	Name  string          `json:"name"`
	Items []CheckListItem `json:"items"`
}

type TaskValidator interface {
	ValidateTask(ctx context.Context, task *stdlib.Reference[*Task]) error
}

type TaskValidatorNode interface {
	psi.Node

	TaskValidator
}

type TaskDescription struct {
	psi.NodeBase

	Description string                                 `json:"description"`
	CheckList   *stdlib.Reference[*CheckList]          `json:"check_list"`
	Validators  []*stdlib.Reference[TaskValidatorNode] `json:"validators"`
}

type TaskLogOperation string

const (
	TaskLogOperationDelivery TaskLogOperation = "delivery"
	TaskLogOperationFeedback TaskLogOperation = "feedback"
	TaskLogOperationReward   TaskLogOperation = "reward"
)

type TaskLogEntry struct {
	psi.NodeBase

	Op TaskLogOperation `json:"op"`
}

type TaskLog struct {
	psi.NodeBase
}

type TaskDirector struct {
	psi.NodeBase

	Name string                             `json:"name"`
	Goal stdlib.Reference[*TaskDescription] `json:"goal"`
}
