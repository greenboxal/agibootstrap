package platform

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/project/filetypes"
)

var Module = fx.Module(
	"platform",

	filetypes.Module,

	fx.Provide(Instance),
)
