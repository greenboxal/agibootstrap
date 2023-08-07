package ws

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
)

var Module = fx.Module(
	"apis/ws",

	apimachinery.ProvideHttpService[*Handler]("/ws/v1", NewHandler),
)
