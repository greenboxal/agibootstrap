package psidb

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/psidb/apis/rest"
	"github.com/greenboxal/agibootstrap/psidb/apis/rpc"
	rpcv1 "github.com/greenboxal/agibootstrap/psidb/apis/rpc/v1"
	"github.com/greenboxal/agibootstrap/psidb/core"
	"github.com/greenboxal/agibootstrap/psidb/modules"
)

var BaseModules = fx.Options(
	logging.Module,
	apimachinery.Module,
	core.Module,
	modules.Module,
	rest.Module,
	rpc.Module,
	rpcv1.Module,
)
