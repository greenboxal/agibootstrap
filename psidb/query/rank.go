package query

import (
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type ComparatorFunc func(ctx QueryContext, a, b psi.Node) int

type Rank struct {
	Comparator ComparatorFunc `json:"comparator"`
	Limit      int            `json:"limit"`
}

func (r Rank) Run(ctx QueryContext, in Iterator) (Iterator, error) {
	capHint := r.Limit

	if capHint == -1 {
		capHint = in.Len()
	}

	if capHint == -1 {
		capHint = 2
	}

	ch := make(chan IteratorItem, capHint)

	go func() {
		defer close(ch)

		result := make([]psi.Node, 0, capHint)

		for in.Next() {
			result = append(result, in.Value())

			idx, _ := slices.BinarySearchFunc(result, in.Value(), func(i, j psi.Node) int {
				return r.Comparator(ctx, i, j)
			})

			if idx >= cap(result) {
				continue
			}

			result = slices.Insert(result, idx, in.Value())

			if r.Limit != -1 && len(result) > r.Limit {
				result = result[:r.Limit]
			}
		}

		if in.Err() != nil {
			ch <- IteratorItem{Err: in.Err()}
			return
		}

		for _, n := range result {
			ch <- IteratorItem{Value: n}
		}
	}()

	return NewChanIterator(ch), nil
}
