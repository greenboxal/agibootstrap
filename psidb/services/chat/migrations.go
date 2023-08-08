package chat

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/services/migrations"
)

var chatsBasePath = psi.MustParsePath("//_Chats")

var migrationSet = migrations.NewOrderedMigrationSet(
	"chat",

	// Create the topic type
	migrations.Migration{
		Name: "create-topics-type",

		Up: func(ctx context.Context, tx coreapi.Transaction) error {
			_, err := psi.ResolveOrCreate(ctx, tx.Graph(), chatsBasePath, func() *stdlib.Collection {
				return stdlib.NewCollection("_Chats")
			})

			if err != nil {
				return err
			}

			return nil
		},
	},
)
