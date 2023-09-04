package psidsadapter

import (
	"context"

	"github.com/ipfs/go-datastore"
)

var dsBatchCtxKey = &struct{ dsBatchCtxKey string }{dsBatchCtxKey: "dsBatchCtxKey"}

func WithBatch(ctx context.Context, batch datastore.Batch) context.Context {
	return context.WithValue(ctx, dsBatchCtxKey, batch)
}

func GetBatch(ctx context.Context) datastore.Batch {
	if batch, ok := ctx.Value(dsBatchCtxKey).(datastore.Batch); ok {
		return batch
	}

	return nil
}

func GetBatchWriter(ctx context.Context, ds datastore.Batching) datastore.Write {
	if b := GetBatch(ctx); b != nil {
		return b
	}

	return ds
}
