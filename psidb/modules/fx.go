package modules

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/modules/vfs"
)

var Module = fx.Module(
	"modules",

	vfs.Module,
	gpt.Module,
)
