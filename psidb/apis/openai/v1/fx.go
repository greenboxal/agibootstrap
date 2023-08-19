package openaiv1

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
)

var Module = fx.Module(
	"apis/openai",

	apimachinery.ProvideHttpService[*Router]("/v1/openai", NewRouter),
)
