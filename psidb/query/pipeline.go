package query

type Pipeline struct {
	Queries []Query `json:"queries"`
}

func (p Pipeline) Run(ctx QueryContext, in Iterator) (Iterator, error) {
	for _, q := range p.Queries {
		result, err := q.Run(ctx, in)

		if err != nil {
			return nil, err
		}

		in = result
	}

	return in, nil
}
