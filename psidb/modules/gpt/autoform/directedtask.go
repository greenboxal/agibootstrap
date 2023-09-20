package autoform

import (
	"context"

	"github.com/go-errors/errors"
)

type DirectedTaskHandler interface {
	OnPrepare(ctx context.Context) error
	OnStep(ctx context.Context) error

	CheckComplete(ctx context.Context) (bool, error)
	HandleError(err error) error
}

type DirectedTask struct {
	handler DirectedTaskHandler

	step     int
	complete bool
}

func NewDirectedTask(handler DirectedTaskHandler) *DirectedTask {
	return &DirectedTask{
		handler: handler,
	}
}

func (t *DirectedTask) GetCurrentStep() int     { return t.step }
func (t *DirectedTask) SetCurrentStep(step int) { t.step = step }
func (t *DirectedTask) IsComplete() bool        { return t.complete }

func (t *DirectedTask) Step(ctx context.Context) (err error) {
	if t.complete {
		return nil
	}

	if t.step == 0 {
		if err := t.handler.OnPrepare(ctx); err != nil {
			return err
		}
	}

	t.step++

	defer func() {
		if e := recover(); e != nil {
			err = errors.Wrap(e, 1)
		}

		if err != nil {
			err = t.handler.HandleError(err)
		}
	}()

	if err := t.handler.OnStep(ctx); err != nil {
		return err
	}

	t.complete, err = t.handler.CheckComplete(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (t *DirectedTask) Run(ctx context.Context, steps int) (bool, error) {
	for i := 0; i < steps; i++ {
		if err := t.Step(ctx); err != nil {
			return false, err
		}

		if t.complete {
			return true, nil
		}
	}

	return false, nil
}

func (t *DirectedTask) RunToCompletion(ctx context.Context) (bool, error) {
	for !t.complete {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		if err := t.Step(ctx); err != nil {
			return false, err
		}
	}

	return true, nil
}
