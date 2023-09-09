package network

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

func NewPubSub(h host.Host) (*pubsub.PubSub, error) {
	return pubsub.NewGossipSub(context.Background(), h)
}
