package golang

import (
	"github.com/alecthomas/repr"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

var Module = fx.Module(
	"langs/golang",
)

func init() {
	pt := typesystem.GetType[Parser]()
	repr.Println(pt.Name().MangledName())
}
