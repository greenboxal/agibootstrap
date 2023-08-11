package kb

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
)

var logger = logging.GetLogger("kb")

var Module = fx.Module(
	"services/kb",
)
