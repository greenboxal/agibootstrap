package coreapi

import (
	"github.com/dgraph-io/badger"
	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime/linking"
)

type DataStore interface {
	datastore.Batching

	DB() *badger.DB
	LinkSystem() *linking.LinkSystem
}
