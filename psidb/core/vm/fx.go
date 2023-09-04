package vm

import (
	`go.uber.org/fx`

	`github.com/greenboxal/agibootstrap/pkg/platform/inject`
)

var FXModule = fx.Module(
	"vm",

	fx.Provide(NewVM),
	inject.WithRegisteredService[*VM](inject.ServiceRegistrationScopeSingleton),
)
