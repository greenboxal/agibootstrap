package thoughtstream

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
)

type Iterator = iterators.Iterator[*Thought]

type Stream interface {
	Iterator

	Previous() bool

	Pointer() Pointer
	Thought() *Thought

	LA(n int) *Thought
	Peek() *Thought
	Seek(p Pointer) error
	Slice(from, to Pointer) Stream

	Fork(options ...ForkOption) MutableStream
}

type MutableStream interface {
	Stream

	Append(t *Thought)

	Merge(t Stream, options ...MergeOption)
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
