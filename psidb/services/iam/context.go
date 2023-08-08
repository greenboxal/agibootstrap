package iam

import "context"

var ctxKeyIdentity = &struct{}{}

func WithIdentity(ctx context.Context, identity *Identity) context.Context {
	return context.WithValue(ctx, ctxKeyIdentity, identity)
}

func IdentityFromContext(ctx context.Context) *Identity {
	id := ctx.Value(ctxKeyIdentity)

	if id == nil {
		return nil
	}

	return id.(*Identity)
}
