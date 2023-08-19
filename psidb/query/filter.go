package query

import "github.com/greenboxal/agibootstrap/pkg/psi"

type FilterFunc func(ctx QueryContext, n psi.Node) bool

type Filter struct {
	Filter FilterFunc `json:"filter"`
}

func (f Filter) Run(ctx QueryContext, in Iterator) (Iterator, error) {
	return NewIterator(func() (psi.Node, bool, error) {
		for in.Next() {
			if f.Filter(ctx, in.Value()) {
				return in.Value(), true, nil
			}
		}

		return nil, false, nil
	}), nil
}
