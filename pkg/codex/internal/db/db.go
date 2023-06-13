package db

import (
	"context"

	"github.com/ipfs/go-datastore"
)

type DB interface {
	Get(ctx context.Context, key Key) ([]byte, error)
	Put(ctx context.Context, key Key, value []byte) error
}

// db implements the database interface for the codex project internal database.
type db struct {
	ds datastore.Batching
}

type Key []string

// New creates a new database instance.
func New(ds datastore.Batching) DB {
	return &db{
		ds: ds,
	}
}

func (db *db) Get(ctx context.Context, key Key) ([]byte, error) {
	return db.ds.Get(ctx, datastore.KeyWithNamespaces(key))
}

func (db *db) Put(ctx context.Context, key Key, value []byte) error {
	return db.ds.Put(ctx, datastore.KeyWithNamespaces(key), value)
}
