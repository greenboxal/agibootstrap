package thoughtstream

import (
	"time"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/stdlib"
)

type ThoughtAddress = psi.Path

type Pointer struct {
	Parent    ThoughtAddress
	Level     int
	Clock     int
	Timestamp time.Time
}

type Stream interface {
	stdlib.Iterator[Thought]

	Pointer() Pointer
	Thought() *Thought

	Seek(p Pointer) error
}

type Thread interface {
	Stream

	Append(t *Thought)

	Fork(options ...ForkOption) Thread
	Merge(t Thread, options ...MergeOption)
}

type ForkOptions struct {
}

type ForkOption func(*ForkOptions)

func NewForkOptions(opts ...ForkOption) ForkOptions {
	var fo ForkOptions
	for _, opt := range opts {
		opt(&fo)
	}
	return fo
}

type MergeOptions struct {
}

type MergeOption func(*MergeOptions)

func NewMergeOptions(opts ...MergeOption) MergeOptions {
	var mo MergeOptions
	for _, opt := range opts {
		opt(&mo)
	}
	return mo
}
