package docs

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
)

var Module = fx.Module(
	"services/docs",

	fx.Provide(NewIndexManager),

	inject.WithRegisteredService[*IndexManager](inject.ServiceRegistrationScopeSingleton),
)
