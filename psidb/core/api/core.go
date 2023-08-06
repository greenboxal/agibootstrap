package coreapi

import (
	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime/linking"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

type Core interface {
	Config() *Config
	DataStore() DataStore
	Journal() *graphfs.Journal
	Checkpoint() graphfs.Checkpoint
	LinkSystem() *linking.LinkSystem
	VirtualGraph() *graphfs.VirtualGraph
	ServiceProvider() inject.ServiceProvider

	TransactionOperations
	graphfs.ReplicationManager
}

type DataStore interface {
	datastore.Batching
}
