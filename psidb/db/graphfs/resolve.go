package graphfs

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func Resolve(ctx context.Context, root *CacheEntry, path psi.Path) (*CacheEntry, error) {
	components := path.Components()

	for _, e := range components {
		if root.IsNegative() {
			return root, psi.ErrNodeNotFound
		}

		child, err := root.Lookup(ctx, e)

		if err != nil {
			return nil, err
		}

		root = child
	}

	return root, nil
}
