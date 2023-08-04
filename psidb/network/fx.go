package network

import (
	"context"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"network",

	fx.Provide(NewConnMgr),
	fx.Provide(NewNetwork),
	fx.Provide(NewPubSub),
)

func NewPubSub(h host.Host) (*pubsub.PubSub, error) {
	return pubsub.NewGossipSub(context.Background(), h)
}

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
