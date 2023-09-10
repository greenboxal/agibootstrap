package pubsub

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"services/pubsub",

	fx.Provide(NewManager),
	fx.Provide(NewDispatcher),

	fx.Invoke(func(dispatcher *Dispatcher) {}),
)
