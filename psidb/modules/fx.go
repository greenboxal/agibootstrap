package modules

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/modules/repl"
)

var Module = fx.Module(
	"modules",

	gpt.Module,
	repl.Module,
)
