package query

import (
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Select struct {
	Path psi.Path `json:"path"`
}

func (s Select) Run(ctx QueryContext, in Iterator) (Iterator, error) {
	done := false

	return NewIterator(func() (psi.Node, bool, error) {
		if done {
			return nil, false, nil
		}

		n, err := ctx.Transaction().Resolve(ctx.Context(), s.Path)

		if errors.Is(err, psi.ErrNodeNotFound) {
			return nil, false, nil
		} else if err != nil {
			return nil, false, err
		}

		return n, true, nil
	}), nil
}
