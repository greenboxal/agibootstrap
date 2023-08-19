package query

import "github.com/greenboxal/agibootstrap/pkg/psi"

type Traverse struct {
	Filter FilterFunc `json:"filter"`
}

func (t Traverse) Run(ctx QueryContext, in Iterator) (Iterator, error) {
	ch := make(chan IteratorItem)

	go func() {
		defer close(ch)

		cursor := psi.NewCursor()

		cursor.Enqueue(in)

		for in.Next() {
			select {
			case <-ctx.Context().Done():
				return
			default:
			}

			err := cursor.Walk(in.Value(), func(c psi.Cursor, entering bool) error {
				select {
				case <-ctx.Context().Done():
					return psi.ErrAbort
				default:
				}

				if !entering {
					return nil
				}

				if t.Filter != nil && !t.Filter(ctx, c.Value()) {
					ch <- IteratorItem{Value: c.Value()}
				}

				return nil
			})

			if err != nil {
				ch <- IteratorItem{Err: err}
				return
			}
		}

		if in.Err() != nil {
			ch <- IteratorItem{Err: in.Err()}
			return
		}
	}()

	return NewChanIterator(ch), nil
}
