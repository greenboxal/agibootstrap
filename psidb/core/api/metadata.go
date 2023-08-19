package coreapi

import (
	"github.com/dgraph-io/badger"
	"github.com/ipfs/go-datastore"
)

type DataStore interface {
	datastore.Batching
}

type MetadataStore interface {
	DB() *badger.DB

	DataStore
}
