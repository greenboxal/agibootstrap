package thoughtstream

import (
	"context"

	"github.com/ipfs/go-cid"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
)

type MergeStrategy interface {
	Apply(ctx context.Context, s MutableStream, streams ...Stream) (Iterator, error)
}

type MergeStrategyFunc func(ctx context.Context, s MutableStream, streams ...Stream) (Iterator, error)

func (f MergeStrategyFunc) Apply(ctx context.Context, s MutableStream, streams ...Stream) (Iterator, error) {
	return f(ctx, s, streams...)
}

func FlatTimeMergeStrategy() MergeStrategyFunc {
	return func(ctx context.Context, s MutableStream, streams ...Stream) (Iterator, error) {
		it := iterators.Concat[Stream, *Thought](streams...)

		it = iterators.SortWith(it, func(a, b *Thought) int {
			return a.Pointer.CompareTo(b.Pointer)
		})

		it = iterators.Map(it, func(t *Thought) *Thought {
			t = t.Clone()
			t.Pointer = Pointer{}
			return t
		})

		return it, nil
	}
}

type Resolver interface {
	ResolveThought(ctx context.Context, id cid.Cid) (*Thought, error)
	ResolveBranch(ctx context.Context, id cid.Cid) (Branch, error)
}

func HierarchicalTimeMergeStrategy(r Resolver) MergeStrategyFunc {
	return func(ctx context.Context, s MutableStream, streams ...Stream) (Iterator, error) {
		it := iterators.Concat[Stream, *Thought](streams...)

		it = iterators.SortWith(it, func(a, b *Thought) int {
			var x, y *Thought

			if a.Pointer.Level < b.Pointer.Level {
				x = a
				y = b
			} else {
				x = b
				y = a
			}

			for x != nil && y != nil && !x.Pointer.IsSiblingOf(y.Pointer) {
				xp := x.Pointer
				yp := y.Pointer
				xl := xp.Level
				yl := yp.Level

				if xl == yl {
					x = nil
					y = nil
					break
				}

				if xl > yl {
					tmp := x
					x = y
					y = tmp
					continue
				}

				yid, err := cid.Parse(y.Pointer.Parent)

				if err != nil {
					panic(err)
				}

				p, err := r.ResolveBranch(ctx, yid)

				if err != nil {
					panic(err)
				}

				y = p.Head()
			}

			if x == nil {
				x = a
			}

			if y == nil {
				y = b
			}

			return x.Pointer.CompareTo(y.Pointer)
		})

		it = iterators.Map(it, func(t *Thought) *Thought {
			t = t.Clone()
			t.Pointer = Pointer{}
			return t
		})

		return it, nil
	}
}
