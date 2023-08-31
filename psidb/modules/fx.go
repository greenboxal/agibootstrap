package modules

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
)

var Module = fx.Module(
	"modules",

	gpt.Module,
)
