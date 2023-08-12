package typing

import "go.uber.org/fx"

var Module = fx.Module(
	"services/typing",

	fx.Provide(NewManager),

	fx.Invoke(func(m *Manager) {}),
)
