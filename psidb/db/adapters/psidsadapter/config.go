package psidsadapter

import (
	`context`

	coreapi `github.com/greenboxal/agibootstrap/psidb/core/api`
)

type BadgerSuperBlockConfig struct {
	MetadataStoreConfig coreapi.MetadataStoreConfig
}

func (b BadgerSuperBlockConfig) Mount(ctx context.Context, md coreapi.MountDefinition) (any, error) {
	store, err := b.MetadataStoreConfig.CreateMetadataStore(ctx)

	if err != nil {
		return nil, err
	}

	return NewDataStoreSuperBlock(ctx, store, md.Name, true)
}
