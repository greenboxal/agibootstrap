package coreapi

import (
	"context"
	"time"

	"github.com/ipld/go-ipld-prime/linking"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
)

type SessionManager interface {
	CreateSession() Session
	GetSession(id string) Session
	GetOrCreateSession(id SessionConfig) Session
}

type SessionClient interface {
	SendSessionMessage(sessionId string, msg SessionMessage) error
}

type Session interface {
	UUID() string

	KeepAlive()
	LastKeepAlive() time.Time
	Closing() <-chan struct{}
	Closed() <-chan struct{}
	Ready() <-chan struct{}
	Err() error

	Journal() Journal
	MetadataStore() DataStore
	VirtualGraph() VirtualGraph
	LinkSystem() *linking.LinkSystem
	ServiceProvider() inject.ServiceProvider

	AttachClient(client SessionClient)
	DetachClient(client SessionClient)

	ReceiveMessage(m SessionMessage)
	SendMessage(m SessionMessage)

	ShutdownNow()
	ShutdownAndWait(ctx context.Context) error
	Close() error

	TransactionOperations
}

var ctxKeySession = struct{ name string }{name: "PsiDbSession"}

func WithSession(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, ctxKeySession, session)
}

func GetSession(ctx context.Context) Session {
	v, ok := ctx.Value(ctxKeySession).(Session)

	if !ok {
		return nil
	}

	return v
}
