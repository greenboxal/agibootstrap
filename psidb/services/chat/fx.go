package chat

import "go.uber.org/fx"

var Module = fx.Module(
	"services/chat",

	fx.Provide(NewService),
)
