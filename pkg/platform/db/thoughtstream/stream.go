package thoughtstream

import (
	"context"

	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
)

type Iterable = iterators.Iterable[*Thought]
type Iterator = iterators.Iterator[*Thought]

type Stream interface {
	Iterator
	Iterable

	Previous() bool
	Reversed() Stream

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

type hierarchicalStream struct {
	stack []Branch

	branch Branch
	stream Stream

	xmin Pointer
	xmax Pointer

	resolver Resolver
}

func NewHierarchicalStream(branch Branch) Stream {
	h := &hierarchicalStream{
		branch: branch,

		xmin: RootPointer(),
		xmax: Head(),
	}

	return h
}

func (h *hierarchicalStream) Pointer() Pointer   { return h.stream.Pointer() }
func (h *hierarchicalStream) Thought() *Thought  { return h.stream.Thought() }
func (h *hierarchicalStream) Value() *Thought    { return h.stream.Value() }
func (h *hierarchicalStream) Iterator() Iterator { return h.Clone() }

func (h *hierarchicalStream) Reversed() Stream {
	return reversedIterator{
		Stream: h.Clone(),
	}
}

func (h *hierarchicalStream) Next() bool {
	for h.stream == nil || !h.stream.Next() {
		if !h.NextBranch() {
			return false
		}
	}

	return true
}

func (h *hierarchicalStream) Previous() bool {
	for h.stream == nil || !h.stream.Previous() {
		if !h.PreviousBranch() {
			return false
		}
	}

	return true
}

func (h *hierarchicalStream) NextBranch() bool {
	if len(h.stack) == 0 {
		return false
	}

	h.branch = h.stack[len(h.stack)-1]
	h.stack = h.stack[:len(h.stack)-1]
	h.stream = h.branch.Stream().Slice(h.xmin, h.xmax)

	if err := h.stream.Seek(RootPointer()); err != nil {
		panic(err)
	}

	return true
}

func (h *hierarchicalStream) PreviousBranch() bool {
	if h.branch == nil {
		return false
	}

	parent := h.branch.ParentBranch()

	if parent == nil {
		return false
	}

	h.stack = append(h.stack, h.branch)
	h.branch = parent
	h.stream = h.branch.Stream().Slice(h.xmin, h.xmax)

	if err := h.stream.Seek(Head()); err != nil {
		panic(err)
	}

	return true
}

func (h *hierarchicalStream) LA(n int) *Thought { return h.stream.LA(n) }
func (h *hierarchicalStream) Peek() *Thought    { return h.LA(0) }

func (h *hierarchicalStream) Seek(p Pointer) error {
	if p.Less(h.xmin) {
		p = h.xmin
	}

	if !h.xmax.IsHead() {
		if !p.Less(h.xmax) {
			p = h.xmax
		}
	}

	for p.Less(h.branch.BasePointer()) {
		if !h.PreviousBranch() {
			break
		}
	}

	return h.stream.Seek(p)
}

func (h *hierarchicalStream) Slice(from, to Pointer) Stream {
	if from.Less(h.xmin) {
		from = h.xmin
	}

	if !h.xmax.IsHead() {
		if !to.Less(h.xmax) {
			to = h.xmax
		}
	}

	return &hierarchicalStream{
		branch: h.branch,
		stream: h.stream.Slice(from, to),
		xmin:   from,
		xmax:   to,
	}
}

func (h *hierarchicalStream) Fork(options ...ForkOption) MutableStream {
	return h.stream.Fork(options...)
}

func (h *hierarchicalStream) AsBranch() Branch { return h.stream.AsBranch() }

func (h *hierarchicalStream) Clone() Stream {
	clone := &hierarchicalStream{
		stack:  slices.Clone(h.stack),
		branch: h.branch,

		xmin: h.xmin,
		xmax: h.xmax,
	}

	if h.stream != nil {
		clone.stream = clone.branch.Stream()

		if err := clone.stream.Seek(h.stream.Pointer()); err != nil {
			panic(err)
		}
	}

	return clone
}

type reversedIterator struct {
	Stream
}

func (r reversedIterator) Iterator() Iterator { return r }
func (r reversedIterator) Next() bool         { return r.Stream.Previous() }
func (r reversedIterator) Previous() bool     { return r.Stream.Next() }
