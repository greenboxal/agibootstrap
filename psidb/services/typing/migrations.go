package typing

import (
	"context"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/migrations"
)

var RootPath = psi.MustParsePath("//_TypeSystem")

var migrationSet = migrations.NewOrderedMigrationSet(
	"typing",

	// Create the topic type
	migrations.Migration{
		Name: "create-root",

		Up: func(ctx context.Context, tx coreapi.Transaction) error {
			_, err := psi.ResolveOrCreate(ctx, tx.Graph(), RootPath, func() *stdlib.Collection {
				return stdlib.NewCollection(RootPath.Name().Name)
			})

			if err != nil {
				return err
			}

			return nil
		},
	},
)
