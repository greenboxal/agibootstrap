package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
)

type AgentContext interface {
	Context() context.Context
	Profile() *Profile
	Agent() Agent
	Branch() thoughtdb.Branch
	Stream() thoughtdb.Cursor
	WorldState() WorldState
}

type Agent interface {
	AnalysisSession

	Profile() *Profile
	Log() thoughtdb.Branch
	History() []*thoughtdb.Thought
	WorldState() WorldState

	AttachTo(r Router)
	ForkSession() (AnalysisSession, error)

	Step(ctx context.Context, options ...StepOption) error
}

type StepOptions struct {
	Base   thoughtdb.Branch
	Head   thoughtdb.Pointer
	Stream thoughtdb.Cursor
}

func (opts *StepOptions) Apply(options ...StepOption) error {
	for _, opt := range options {
		if err := opt(opts); err != nil {
			return err
		}
	}

	if opts.Stream == nil {
		opts.Stream = opts.Base.Cursor()
	}

	if !opts.Head.IsZero() {
		if err := opts.Stream.PushPointer(opts.Head); err != nil {
			return err
		}
	}

	return nil
}

type StepOption func(*StepOptions) error

func NewStepOptions(options ...StepOption) (opts StepOptions, err error) {
	if err := opts.Apply(options...); err != nil {
		return opts, err
	}

	return
}

func WithStream(stream thoughtdb.Cursor) StepOption {
	return func(opts *StepOptions) error {
		opts.Stream = stream

		return nil
	}
}

func WithBranch(branch thoughtdb.Branch) StepOption {
	return func(opts *StepOptions) error {
		opts.Base = branch

		return nil
	}
}

func WithHeadPointer(head thoughtdb.Pointer) StepOption {
	return func(opts *StepOptions) error {
		opts.Head = head

		return nil
	}
}
