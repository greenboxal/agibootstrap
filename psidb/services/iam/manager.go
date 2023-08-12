package iam

import (
	"context"

	"go.uber.org/fx"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/services/migrations"
)

type Manager struct {
	core     coreapi.Core
	migrator migrations.Migrator
}

func NewManager(
	lc fx.Lifecycle,
	core coreapi.Core,
	migrator migrations.Migrator,
) *Manager {
	m := &Manager{
		core:     core,
		migrator: migrator,
	}

	lc.Append(fx.Hook{
		OnStart: m.Start,
	})

	return m
}

func (m *Manager) Start(ctx context.Context) error {
	return m.migrator.Migrate(ctx, migrationSet)
}
