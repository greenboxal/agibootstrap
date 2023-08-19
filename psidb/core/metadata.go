package core

import (
	"context"
	"os"
	"path"

	badger2 "github.com/dgraph-io/badger"
	badger "github.com/ipfs/go-ds-badger"
	"go.uber.org/fx"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

func NewDataStore(
	lc fx.Lifecycle,
	cfg *coreapi.Config,
) (coreapi.DataStore, error) {
	dsOpts := badger.DefaultOptions
	dsPath := path.Join(cfg.DataDir, "data")

	if err := os.MkdirAll(dsPath, 0755); err != nil {
		return nil, err
	}

	ds, err := badger.NewDatastore(dsPath, &dsOpts)

	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return ds.Close()
		},
	})

	return ds, nil
}

func NewMetadataStore(db coreapi.DataStore) coreapi.MetadataStore {
	return &MetadataStore{DataStore: db}
}

type MetadataStore struct {
	coreapi.DataStore
}

func (m *MetadataStore) DB() *badger2.DB {
	return m.DataStore.(*badger.Datastore).DB
}
