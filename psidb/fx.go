package psidb

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	mgmtv1 `github.com/greenboxal/agibootstrap/psidb/apis/mgmt/v1`
	openaiv1 `github.com/greenboxal/agibootstrap/psidb/apis/openai/v1`
	restv1 "github.com/greenboxal/agibootstrap/psidb/apis/rest/v1"
	"github.com/greenboxal/agibootstrap/psidb/apis/rpc"
	rpcv1 "github.com/greenboxal/agibootstrap/psidb/apis/rpc/v1"
	"github.com/greenboxal/agibootstrap/psidb/apis/ws"
	"github.com/greenboxal/agibootstrap/psidb/core"
	"github.com/greenboxal/agibootstrap/psidb/modules"
	"github.com/greenboxal/agibootstrap/psidb/services/agents"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
	"github.com/greenboxal/agibootstrap/psidb/services/iam"
	"github.com/greenboxal/agibootstrap/psidb/services/indexing"
	"github.com/greenboxal/agibootstrap/psidb/services/jobs"
	"github.com/greenboxal/agibootstrap/psidb/services/kb"
	"github.com/greenboxal/agibootstrap/psidb/services/migrations"
	"github.com/greenboxal/agibootstrap/psidb/services/pubsub"
	"github.com/greenboxal/agibootstrap/psidb/services/search"
	"github.com/greenboxal/agibootstrap/psidb/services/timekeep"
	"github.com/greenboxal/agibootstrap/psidb/services/typing"
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
	indexing.Module,
	migrations.Module,
	modules.Module,
	search.Module,
	pubsub.Module,
	timekeep.Module,
	typing.Module,
	kb.Module,
	iam.Module,
	jobs.Module,
	chat.Module,
	agents.Module,
	ws.Module,
	rpc.Module,
	rpcv1.Module,
	restv1.Module,
	openaiv1.Module,
	mgmtv1.Module,
)
