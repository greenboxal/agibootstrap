package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
)

type AgentContext interface {
	Context() context.Context
	Profile() *Profile
	Agent() Agent
	Branch() thoughtstream.Branch
	Stream() thoughtstream.Stream
	WorldState() WorldState
}

type Agent interface {
	AnalysisSession

	Profile() *Profile
	Log() thoughtstream.Branch
	History() []*thoughtstream.Thought
	WorldState() WorldState

	AttachTo(r Router)
	ForkSession() (AnalysisSession, error)

	Step(ctx context.Context, options ...StepOption) error
}

type StepOptions struct {
	Base   thoughtstream.Branch
	Head   thoughtstream.Pointer
	Stream thoughtstream.Stream
}

func (opts *StepOptions) Apply(options ...StepOption) error {
	for _, opt := range options {
		if err := opt(opts); err != nil {
			return err
		}
	}

	if opts.Stream == nil {
		opts.Stream = opts.Base.Stream()
	}

	if !opts.Head.IsZero() {
		if err := opts.Stream.Seek(opts.Head); err != nil {
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

func WithStream(stream thoughtstream.Stream) StepOption {
	return func(opts *StepOptions) error {
		opts.Stream = stream

		return nil
	}
}

func WithBranch(branch thoughtstream.Branch) StepOption {
	return func(opts *StepOptions) error {
		opts.Base = branch

		return nil
	}
}

func WithHeadPointer(head thoughtstream.Pointer) StepOption {
	return func(opts *StepOptions) error {
		opts.Head = head

		return nil
	}
}
