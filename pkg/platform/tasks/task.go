package tasks

import (
	"context"

	"github.com/google/uuid"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type TaskProgress interface {
	Context() context.Context
	Update(current, total int)
}

type TaskHandle interface {
	Context() context.Context

	Cancel()
	Wait()
	Err() error
}

type TaskFunc func(progress TaskProgress) error

func (f TaskFunc) Run(progress TaskProgress) error {
	return f(progress)
}

type Task interface {
	psi.Node
	Name() string
	Description() string
	Progress() float64

	IsCompleted() bool
	Error() error

	Done() <-chan struct{}
}

type task struct {
	psi.NodeBase

	tc *taskContext

	uuid        string
	name        string
	description string
}

var TaskType = psi.DefineNodeType[*task](psi.WithRuntimeOnly())

func (t *task) PsiNodeName() string   { return t.uuid }
func (t *task) Name() string          { return t.name }
func (t *task) Description() string   { return t.description }
func (t *task) Progress() float64     { return float64(t.tc.current) / float64(t.tc.total) }
func (t *task) IsCompleted() bool     { return t.tc.complete }
func (t *task) Error() error          { return t.tc.Err() }
func (t *task) Done() <-chan struct{} { return t.tc.done }

func (t *task) Init() {
	t.uuid = uuid.New().String()

	t.NodeBase.Init(t, psi.WithNodeType(TaskType))
}
