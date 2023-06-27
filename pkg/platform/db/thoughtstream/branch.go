package thoughtstream

import (
	"golang.org/x/exp/slices"
)

type Branch interface {
	Base() Branch

	BasePointer() Pointer
	HeadPointer() Pointer

	Interval() Interval

	Stream() Stream
}

func NewBranchFromSlice(base Branch, basePtr Pointer, items ...*Thought) Branch {
	return &sliceBranch{
		base:    base,
		basePtr: basePtr,
		items:   items,
	}
}

type sliceBranch struct {
	base    Branch
	basePtr Pointer
	items   []*Thought
}

var _ Branch = (*sliceBranch)(nil)

func (s *sliceBranch) Base() Branch         { return s.base }
func (s *sliceBranch) BasePointer() Pointer { return s.basePtr }

func (s *sliceBranch) HeadPointer() Pointer {
	if len(s.items) == 0 {
		if s.base != nil {
			return s.base.HeadPointer()
		}

		return Pointer{}
	}

	return s.items[len(s.items)-1].Pointer
}

func (s *sliceBranch) Interval() Interval {
	return Interval{Start: s.BasePointer(), End: s.HeadPointer()}
}

func (s *sliceBranch) Stream() Stream {
	return &branchStream{
		b:      s,
		xmin:   0,
		xmax:   -1,
		offset: 0,
	}
}

type branchStream struct {
	b       *sliceBranch
	xmin    int
	xmax    int
	offset  int
	current *Thought
}

func (s *branchStream) Index() int        { return s.xmin + s.offset }
func (s *branchStream) Value() *Thought   { return s.current }
func (s *branchStream) Thought() *Thought { return s.current }
func (s *branchStream) Pointer() Pointer  { return s.current.Pointer }

func (s *branchStream) Append(t *Thought) {
	if !t.Pointer.IsZero() {
		panic("cannot append a thought with a non-zero pointer")
	}

	t.Pointer = s.Pointer().Next()

	s.b.items = slices.Insert(s.b.items, s.Index(), t)
}

func (s *branchStream) Next() bool {
	index := s.Index()

	if index >= len(s.b.items) || (s.xmax != -1 && index >= s.xmax) {
		return false
	}

	s.current = s.b.items[index]
	s.offset++

	return true
}

func (s *branchStream) Previous() bool {
	index := s.Index()

	if index <= 0 || index <= s.xmin {
		return false
	}

	s.offset--
	s.current = s.b.items[index]

	return true
}

func (s *branchStream) Peek() *Thought {
	return s.LA(0)
}

func (s *branchStream) LA(n int) *Thought {
	index := s.Index() + n

	if index < 0 || index >= len(s.b.items) {
		return nil
	}

	if index < s.xmin || (s.xmax != -1 && index >= s.xmax) {
		return nil
	}

	return s.b.items[index]
}

func (s *branchStream) Seek(p Pointer) error {
	offset, _ := s.binarySearch(p)

	s.offset = offset

	return nil
}

func (s *branchStream) Slice(from, to Pointer) Stream {
	fromIndex, _ := s.binarySearch(from)
	toIndex, _ := s.binarySearch(to)

	return &branchStream{
		b:      s.b,
		xmin:   s.xmin + fromIndex,
		xmax:   s.xmax + toIndex,
		offset: 0,
	}
}

func (s *branchStream) Reset() {
	s.offset = 0
	s.current = nil
}

func (s *branchStream) AsBranch() Branch {
	return &sliceBranch{
		base:    s.b.base,
		basePtr: s.b.basePtr,
		items:   slices.Clone(s.b.items[s.xmin:s.xmax]),
	}
}

func (s *branchStream) binarySearch(p Pointer) (int, bool) {
	return slices.BinarySearchFunc(s.b.items[s.xmin:s.xmax], p, func(a *Thought, b Pointer) int {
		return a.Pointer.CompareTo(b)
	})
}

func (s *branchStream) Fork(options ...ForkOption) MutableStream {
	base := s.b.HeadPointer()

	forked := &sliceBranch{
		base:    s.b,
		basePtr: base,
	}

	return &branchStream{
		b:       forked,
		xmin:    0,
		xmax:    -1,
		offset:  0,
		current: s.current,
	}
}

func (s *branchStream) Merge(t Stream, options ...MergeOption) {

}
