package restv1

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
)

var Module = fx.Module(
	"apis/rest",

	apimachinery.ProvideHttpService[*ResourceHandler]("/v1/psi", NewResourceHandler),
	apimachinery.ProvideHttpService[*RenderHandler]("/v1/render", NewRenderHandler),
	apimachinery.ProvideHttpService[*SearchHandler]("/v1/search", NewSearchHandler),
	apimachinery.ProvideHttpService[*ChatHandler]("/v1/chat", NewChatHandler),
)
