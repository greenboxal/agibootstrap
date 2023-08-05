package psidb

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

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

	fx.WithLogger(func(l *zap.Logger) fxevent.Logger {
		zl := &fxevent.ZapLogger{Logger: l}
		zl.UseLogLevel(-2)
		zl.UseErrorLevel(zap.ErrorLevel)
		return zl
	}),

	apimachinery.Module,
	core.Module,
	modules.Module,
	rest.Module,
	rpc.Module,
	rpcv1.Module,
)
