package rest

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
)

var Module = fx.Module(
	"apis/rest",

	fx.Provide(NewResourceHandler),
	fx.Provide(NewSearchHandler),
	fx.Provide(NewRenderHandler),

	apimachinery.ProvideHttpService[*Router]("/v1", NewRouter),
)
