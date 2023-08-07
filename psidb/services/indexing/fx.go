package indexing

import "go.uber.org/fx"

var Module = fx.Module(
	"indexing",

	fx.Provide(NewIndexManager),
)
