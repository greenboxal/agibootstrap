package vm

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
)

var logger = logging.GetLogger("vm")

var FXModule = fx.Module(
	"vm",

	fx.Provide(NewSupervisor),
	inject.WithRegisteredService[*Supervisor](inject.ServiceRegistrationScopeSingleton),

	fx.Invoke(func(sup *Supervisor) {
		sup.NewIsolate()
		sup.NewIsolate()
	}),
)
