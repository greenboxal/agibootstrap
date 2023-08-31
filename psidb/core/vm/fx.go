package vm

import "go.uber.org/fx"

var FXModule = fx.Module(
	"vm",

	fx.Provide(NewVM),
)
