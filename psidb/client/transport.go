package client

import "context"

type Transport interface {
	Connect(ctx context.Context, incomingCh chan Message) (Connection, error)
}

type Connection interface {
	IsConnected() bool

	SendMessage(ctx context.Context, msg Message) error
	Close() error
}
