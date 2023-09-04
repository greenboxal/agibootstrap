package typing

import (
	`go.uber.org/fx`

	`github.com/greenboxal/agibootstrap/pkg/platform/inject`
)

var Module = fx.Module(
	"services/typing",

	fx.Provide(NewManager),

	inject.WithRegisteredService[*Manager](inject.ServiceRegistrationScopeSingleton),
)
