package mgmtv1

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
)

var Module = fx.Module(
	"apis/mgmt/v1",

	apimachinery.ProvideHttpService[*Router]("/management/v1", NewRouter),
)
