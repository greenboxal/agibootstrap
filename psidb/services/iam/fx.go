package iam

import "go.uber.org/fx"

var Module = fx.Module(
	"services/iam",

	fx.Provide(NewManager),
)
