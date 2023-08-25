package rpcv1

import (
	"go.opentelemetry.io/otel"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/psidb/apis/rpc"
)

var tracer = otel.Tracer("rpcv1")

var Module = fx.Module(
	"apis/rpc/v1",

	rpc.ProvideRpcService[*JournalService](NewJournalService, "Journal"),
	rpc.ProvideRpcService[*NodeService](NewNodeService, "NodeService"),
	rpc.ProvideRpcService[*ObjectStore](NewObjectStore, "ObjectStore"),
	rpc.ProvideRpcService[*Search](NewSearch, "Search"),
	rpc.ProvideRpcService[*Chat](NewChat, "Chat"),
)
