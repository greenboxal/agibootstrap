package filetypes

import "go.uber.org/fx"

var Module = fx.Module(
	"project/filetypes",

	fx.Provide(NewRegistry),
)
