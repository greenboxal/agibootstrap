package vfs

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
)

type Manager = vfs.Manager

var Module = fx.Module(
	"modules/vfs",

	fx.Provide(vfs.NewManager),

	inject.WithRegisteredService[*vfs.Manager](inject.ServiceRegistrationScopeSingleton),
)
