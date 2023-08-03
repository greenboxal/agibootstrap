package rpcv1

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/psidb/apis/rpc"
)

var Module = fx.Module(
	"apis/rpc/v1",

	rpc.ProvideRpcService[*NodeService](NewNodeService, "NodeService"),
	rpc.ProvideRpcService[*ObjectStore](NewObjectStore, "ObjectStore"),
	rpc.ProvideRpcService[*Search](NewSearch, "Search"),
)
