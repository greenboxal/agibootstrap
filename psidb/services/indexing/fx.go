package indexing

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
)

var Module = fx.Module(
	"indexing",

	fx.Provide(NewIndexManager),

	inject.WithRegisteredService[*Manager](inject.ServiceRegistrationScopeSingleton),
)
