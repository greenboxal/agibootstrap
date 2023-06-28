package thoughtstream

import (
	"context"

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

	AsBranch() Branch
}

type MutableStream interface {
	Stream

	Append(t *Thought)

	Merge(ctx context.Context, r Resolver, strategy MergeStrategy, branches ...Branch) error
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
