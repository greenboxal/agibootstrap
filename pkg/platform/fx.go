package platform

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"platform",

	fx.Provide(Instance),
)
