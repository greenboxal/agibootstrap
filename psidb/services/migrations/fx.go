package migrations

import "go.uber.org/fx"

var Module = fx.Module(
	"services/migrations",

	fx.Provide(NewManager),
)
