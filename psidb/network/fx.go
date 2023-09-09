package network

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
)

var logger = logging.GetLogger("network/p2p")

var Module = fx.Module(
	"network",

	fx.Provide(NewConnMgr),
	fx.Provide(NewNetwork),
	fx.Provide(NewPeerStore),
	fx.Provide(NewPubSub),

	fx.Provide(func(n *Network) (res struct {
		fx.Out

		Host host.Host
	}) {
		res.Host = n.host

		return
	}),
)

func NewConnMgr(lc fx.Lifecycle) (*connmgr.BasicConnMgr, error) {
	cm, err := connmgr.NewConnManager(
		100, // Lowwater
		400, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)

	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return cm.Close()
		},
	})

	return cm, nil
}
