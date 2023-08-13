package modules

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/modules/vfs"
	"github.com/greenboxal/agibootstrap/psidb/modules/vm"
)

var Module = fx.Module(
	"modules",

	vfs.Module,
	gpt.Module,
	vm.FXModule,
)
