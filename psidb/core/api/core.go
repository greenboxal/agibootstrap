package coreapi

import (
	`github.com/ipld/go-ipld-prime/linking`

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
)

type Core interface {
	Config() *Config
	Journal() Journal
	MetadataStore() MetadataStore
	VirtualGraph() VirtualGraph
	LinkSystem() *linking.LinkSystem
	ServiceProvider() inject.ServiceProvider

	TransactionOperations
	ReplicationManager
}

type VirtualGraph interface {
	ReplicationOperations
}
