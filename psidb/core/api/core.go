package coreapi

import (
	"github.com/ipld/go-ipld-prime/linking"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
)

type Core interface {
	Config() *Config
	DataStore() DataStore
	Journal() Journal
	Checkpoint() Checkpoint
	LinkSystem() *linking.LinkSystem
	ServiceProvider() inject.ServiceProvider

	TransactionOperations
	ReplicationManager
}
