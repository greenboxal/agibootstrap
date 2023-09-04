package client

import (
	"context"

	"github.com/greenboxal/agibootstrap/psidb/apis/rt/v1"
)

type Transport interface {
	Connect(ctx context.Context, incomingCh chan v1.Message) (Connection, error)
}

type Connection interface {
	IsConnected() bool

	SendMessage(ctx context.Context, msg v1.Message) error
	Close() error
}
