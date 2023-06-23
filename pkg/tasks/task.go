package tasks

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
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
	Name() string
	Description() string

	IsCompleted() bool
	Error() error

	Done() <-chan struct{}
}

type task struct {
	psi.NodeBase

	tc *taskContext

	name        string
	description string
}

func (t *task) Name() string          { return t.name }
func (t *task) Description() string   { return t.description }
func (t *task) IsCompleted() bool     { return t.tc.complete }
func (t *task) Error() error          { return t.tc.err }
func (t *task) Done() <-chan struct{} { return t.tc.done }

func (t *task) Init() {
	t.NodeBase.Init(t, "")
}
