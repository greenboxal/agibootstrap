package coreapi

import (
	"context"

	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime/linking"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
)

type Config struct {
	RootUUID   string
	DataDir    string
	ProjectDir string
}

type DataStore interface {
	datastore.Batching
}

type Core interface {
	Config() *Config
	DataStore() DataStore
	Journal() *graphfs.Journal
	Checkpoint() graphfs.Checkpoint
	LinkSystem() *linking.LinkSystem
	VirtualGraph() *graphfs.VirtualGraph
	ServiceLocator() inject.ServiceLocator

	BeginTransaction(ctx context.Context) (Transaction, error)
	RunTransaction(ctx context.Context, fn func(ctx context.Context, tx Transaction) error) (err error)
}

type GraphOperations interface {
	Add(node psi.Node)
	Remove(node psi.Node)

	Resolve(ctx context.Context, path psi.Path) (psi.Node, error)
}

type Transaction interface {
	GraphOperations

	IsOpen() bool
	Graph() *online.LiveGraph
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
