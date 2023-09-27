package langs

import (
	"go.uber.org/fx"

	golang "github.com/greenboxal/agibootstrap/psidb/services/codeanalysis/langs/go"
)

var Module = fx.Module(
	"langs",

	golang.Module,
)
