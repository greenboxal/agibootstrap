package query

import "github.com/greenboxal/agibootstrap/psidb/psi"

type Merge struct {
	Queries []Query `json:"queries"`
}

func (m Merge) Run(ctx QueryContext, in Iterator) (Iterator, error) {
	var err error

	queries := m.Queries

	return NewIterator(func() (psi.Node, bool, error) {
		for !in.Next() {
			if len(queries) == 0 {
				return nil, false, nil
			}

			in, err = queries[0].Run(ctx, in)

			if err != nil {
				return nil, false, err
			}

			queries = queries[1:]
		}

		return in.Value(), true, in.Err()
	}), nil
}
