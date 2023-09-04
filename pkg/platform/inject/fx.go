package inject

import `go.uber.org/fx`

var Module = fx.Module(
	"platform/inject",

	fx.Provide(NewServiceRegistrationManager),
)
