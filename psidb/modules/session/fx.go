package session

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

var Module = fx.Module(
	"services/session",

	fx.Provide(NewManager),

	fx.Invoke(func(sp inject.ServiceProvider, m coreapi.SessionManager) {
		inject.RegisterInstance(sp, m)
	}),
)
