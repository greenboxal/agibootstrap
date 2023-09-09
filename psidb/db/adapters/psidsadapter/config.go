package psidsadapter

import (
	"context"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type BadgerSuperBlockConfig struct {
	MetadataStoreConfig coreapi.DataStoreConfig
}

func (b BadgerSuperBlockConfig) Mount(ctx context.Context, md coreapi.MountDefinition) (any, error) {
	store, err := b.MetadataStoreConfig.CreateDataStore(ctx)

	if err != nil {
		return nil, err
	}

	return NewDataStoreSuperBlock(ctx, store, md.Name, true)
}
