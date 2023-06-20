package build

import "context"

type StepResult struct {
	ChangeCount int

	Errors []error
}

type Step interface {
	Process(ctx context.Context, bctx *Context) (*StepResult, error)
}
