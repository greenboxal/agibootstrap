package vfs

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
)

var Module = fx.Module(
	"modules/vfs",

	fx.Provide(vfs.NewManager),

	fx.Invoke(func(sp inject.ServiceProvider, m *vfs.Manager) {
		inject.RegisterInstance(sp, m)
	}),
)
