package rpc

import (
	"github.com/greenboxal/aip/aip-sdk/pkg/utils"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
)

var Module = fx.Module(
	"apis/rpc",

	utils.WithBindingRegistry[RpcServiceBinding]("rpc-service-bindings"),

	apimachinery.ProvideHttpService[*RpcService]("/rpc/v1", NewRpcService),
	apimachinery.ProvideHttpService[*Docs]("/rpc/v1/docs", NewDocs),

	fx.Invoke(func(rpcsrv *RpcService, bindings utils.BindingRegistry[RpcServiceBinding]) {
		for _, m := range bindings.Bindings() {
			m.Bind(rpcsrv)
		}
	}),
)
