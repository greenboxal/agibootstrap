package coreapi

import (
	"github.com/ipld/go-ipld-prime/linking"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
)

type Core interface {
	Ready() <-chan struct{}

	Config() *Config
	Journal() Journal
	MetadataStore() DataStore
	VirtualGraph() VirtualGraph
	LinkSystem() *linking.LinkSystem
	ServiceProvider() inject.ServiceProvider

	TransactionOperations
	ReplicationManager
}

type VirtualGraph interface {
	ReplicationOperations
}
