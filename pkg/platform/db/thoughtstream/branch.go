package thoughtstream

import (
	"context"
	"time"

	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Branch interface {
	psi.Node

	Iterable
	ParentBranch() Branch
	Base() *Thought
	Head() *Thought

	ParentPointer() Pointer
	BasePointer() Pointer
	HeadPointer() Pointer

	Interval() Interval

	Stream() Stream
	Mutate() MutableStream

	Fork(options ...ForkOption) MutableStream

	Len() int
}

func NewBranchFromSlice(parent Branch, basePtr Pointer, items ...*Thought) Branch {
	b := &branchImpl{
		parent:  parent,
		basePtr: basePtr,
		items:   items,
	}

	if parent != nil {
		b.parentPtr = parent.BasePointer()
	} else {
		b.parentPtr = RootPointer()
	}

	b.Init(b, "")

	return b
}

type branchImpl struct {
	psi.NodeBase

	parent    Branch
	parentPtr Pointer

	basePtr Pointer

	items []*Thought
}

var _ Branch = (*branchImpl)(nil)

func (s *branchImpl) ParentBranch() Branch   { return s.parent }
func (s *branchImpl) ParentPointer() Pointer { return s.parentPtr }
func (s *branchImpl) BasePointer() Pointer   { return s.basePtr }

func (s *branchImpl) Head() *Thought {
	if len(s.items) == 0 {
		return nil
	}

	return s.items[len(s.items)-1]
}

func (s *branchImpl) Base() *Thought {
	if len(s.items) == 0 {
		return nil
	}
	return s.items[0]
}

func (s *branchImpl) Len() int {
	return len(s.items)
}

func (s *branchImpl) HeadPointer() Pointer {
	if len(s.items) == 0 {
		return s.basePtr
	}

	return s.items[len(s.items)-1].Pointer
}

func (s *branchImpl) Interval() Interval {
	return Interval{Start: s.BasePointer(), End: s.HeadPointer()}
}

func (s *branchImpl) Mutate() MutableStream {
	return &branchStreamImpl{
		b:       s,
		xmin:    0,
		xmax:    -1,
		offset:  len(s.items),
		current: s.Head(),
		pointer: s.HeadPointer(),
	}
}

func (s *branchImpl) Stream() Stream {
	return &branchStreamImpl{
		b:       s,
		xmin:    0,
		xmax:    -1,
		offset:  0,
		current: s.Base(),
		pointer: s.BasePointer(),
	}
}

func (s *branchImpl) Iterator() Iterator {
	return s.Stream()
}

func (s *branchImpl) Fork(options ...ForkOption) MutableStream {
	return s.Mutate().Fork()
}

type branchStreamImpl struct {
	b       *branchImpl
	xmin    int
	xmax    int
	offset  int
	current *Thought
	pointer Pointer
}

func (s *branchStreamImpl) Index() int        { return s.xmin + s.offset }
func (s *branchStreamImpl) Value() *Thought   { return s.current }
func (s *branchStreamImpl) Thought() *Thought { return s.current }
func (s *branchStreamImpl) Pointer() Pointer  { return s.pointer }
func (s *branchStreamImpl) Branch() Branch    { return s.b }

func (s *branchStreamImpl) Iterator() Iterator { return s.Clone() }

func (s *branchStreamImpl) Reversed() Stream {
	return reversedIterator{
		Stream: s.Clone(),
	}
}

func (s *branchStreamImpl) Append(t *Thought) {
	if t.Parent() != nil && t.Parent() != s.b {
		panic("cannot append a thought with a non-nil parent")
	}

	if s.xmax != -1 {
		panic("cannot append to a bounded stream")
	}

	t.SetParent(s.b)

	prev := s.Pointer()

	t.Pointer.Parent = s.b.ParentPointer().Address()
	t.Pointer.Previous = prev.Address()
	t.Pointer.Timestamp = time.Now()
	t.Pointer.Clock = prev.Clock + 1
	t.Pointer.Level = prev.Level

	s.b.items = slices.Insert(s.b.items, s.Index(), t)
	s.current = t
	s.pointer = t.Pointer
	s.offset++
}

func (s *branchStreamImpl) Next() bool {
	index := s.Index()

	if index >= len(s.b.items) || (s.xmax != -1 && index >= s.xmax) {
		return false
	}

	s.current = s.b.items[index]
	s.pointer = s.current.Pointer
	s.offset++

	return true
}

func (s *branchStreamImpl) Previous() bool {
	index := s.Index()

	if index <= 0 || index <= s.xmin {
		return false
	}

	s.offset--
	s.current = s.b.items[index-1]
	s.pointer = s.current.Pointer

	return true
}

func (s *branchStreamImpl) Peek() *Thought {
	return s.LA(0)
}

func (s *branchStreamImpl) LA(n int) *Thought {
	index := s.Index() + n

	if index < 0 || index >= len(s.b.items) {
		return nil
	}

	if index < s.xmin || (s.xmax != -1 && index >= s.xmax) {
		return nil
	}

	return s.b.items[index]
}

func (s *branchStreamImpl) Seek(p Pointer) error {
	if p.IsHead() {
		s.offset = len(s.b.items)
		return nil
	} else if p.IsRoot() {
		s.offset = 0
		return nil
	}

	offset, _ := s.binarySearch(p)

	s.offset = offset

	return nil
}

func (s *branchStreamImpl) Slice(from, to Pointer) Stream {
	fromIndex, _ := s.binarySearch(from)
	toIndex, _ := s.binarySearch(to)

	return &branchStreamImpl{
		b:      s.b,
		xmin:   s.xmin + fromIndex,
		xmax:   s.xmax + toIndex,
		offset: 0,
	}
}

func (s *branchStreamImpl) Reset() {
	s.offset = 0
	s.current = nil
}

func (s *branchStreamImpl) AsBranch() Branch {
	min := s.xmin

	if min < 0 {
		min = 0
	}

	max := s.xmax

	if max < 0 {
		max = len(s.b.items)
	}

	return NewBranchFromSlice(s.b.parent, s.b.parentPtr, s.b.items[min:max]...)
}

func (s *branchStreamImpl) Fork(options ...ForkOption) MutableStream {
	basePtr := s.Pointer()
	basePtr.Level++

	forked := NewBranchFromSlice(s.b, basePtr)

	return forked.Mutate()
}

func (s *branchStreamImpl) Merge(ctx context.Context, r Resolver, strategy MergeStrategy, branches ...Branch) (err error) {
	mergeStreams := make([]Stream, 0, len(branches)+1)

	for _, branch := range branches {
		stream := branch.Stream()
		mergeStreams = append(mergeStreams, stream.Slice(RootPointer(), stream.Pointer()))
	}

	merged, err := strategy.Apply(ctx, s, mergeStreams...)

	if err != nil {
		return err
	}

	for merged.Next() {
		s.Append(merged.Value())
	}

	return nil
}

func (s *branchStreamImpl) Clone() Stream {
	return &branchStreamImpl{
		b:       s.b,
		xmin:    s.xmin,
		xmax:    s.xmax,
		offset:  s.offset,
		current: s.current,
		pointer: s.pointer,
	}
}

func (s *branchStreamImpl) binarySearch(p Pointer) (int, bool) {
	min := s.xmin

	if min < 0 {
		min = 0
	}

	max := s.xmax

	if max < 0 {
		max = len(s.b.items)
	}

	return slices.BinarySearchFunc(s.b.items[min:max], p, func(a *Thought, b Pointer) int {
		return a.Pointer.CompareTo(b)
	})
}
