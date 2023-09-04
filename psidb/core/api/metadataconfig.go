package coreapi

import (
	`context`
	`os`

	badger2 `github.com/dgraph-io/badger`
	badger `github.com/ipfs/go-ds-badger`
)

type MetadataStoreConfig interface {
	CreateMetadataStore(ctx context.Context) (MetadataStore, error)
}

type ExistingMetadataStore struct {
	MetadataStore
}

func (e ExistingMetadataStore) CreateMetadataStore(ctx context.Context) (MetadataStore, error) {
	return e, nil
}

func (e ExistingMetadataStore) Close() error {
	return nil
}

type BadgerMetadataStoreConfig struct {
	Path string `json:"path"`
}

type BadgerMetadataStore struct {
	DataStore
}

func (m *BadgerMetadataStore) DB() *badger2.DB {
	return m.DataStore.(*badger.Datastore).DB
}

func (b BadgerMetadataStoreConfig) CreateMetadataStore(ctx context.Context) (MetadataStore, error) {
	dsOpts := badger.DefaultOptions

	if err := os.MkdirAll(b.Path, 0755); err != nil {
		return nil, err
	}

	ds, err := badger.NewDatastore(b.Path, &dsOpts)

	if err != nil {
		return nil, err
	}

	return &BadgerMetadataStore{DataStore: ds}, nil
}
