package thoughtdb

import "context"

type MergeStrategy interface {
	Merge(ctx context.Context, r *Repo, head Branch, forks []Branch) error
}

type MergeStrategyFunc func(ctx context.Context, r *Repo, head Branch, forks []Branch) error

func (f MergeStrategyFunc) Merge(ctx context.Context, r *Repo, head Branch, forks []Branch) error {
	return f(ctx, r, head, forks)
}

func FlatTimeMergeStrategy() MergeStrategyFunc {
	return func(ctx context.Context, r *Repo, head Branch, forks []Branch) error {
		base := head.Pointer()

		for _, f := range forks {
			for it := f.Cursor().IterateParents(); it.Next(); {
				if it.Value().Pointer == base {
					break
				}

				t := it.Value().Clone()
				t.Pointer = Pointer{}

				if err := head.Commit(ctx, t); err != nil {
					return err
				}
			}
		}

		return nil
	}
}
