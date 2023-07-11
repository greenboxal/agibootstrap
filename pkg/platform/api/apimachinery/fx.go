package apimachinery

import "go.uber.org/fx"

var Module = fx.Module(
	"apimachinery",

	fx.Provide(NewServer),
)
