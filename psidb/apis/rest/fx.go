package rest

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
)

var Module = fx.Module(
	"apis/rest",

	apimachinery.ProvideHttpService[*Handler]("/v1", NewRouter),
)
