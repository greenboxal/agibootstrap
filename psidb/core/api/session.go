package coreapi

import (
	"context"
	"time"
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

	ReceiveMessage(m SessionMessage)
	SendMessage(m SessionMessage)

	AttachClient(client SessionClient)
	DetachClient(client SessionClient)

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
