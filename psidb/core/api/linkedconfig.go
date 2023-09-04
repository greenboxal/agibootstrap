package coreapi

import (
	`context`
	`encoding/hex`

	`github.com/ipld/go-ipld-prime/linking`
	cidlink `github.com/ipld/go-ipld-prime/linking/cid`
	`github.com/ipld/go-ipld-prime/storage/dsadapter`
)

type LinkedStoreConfig interface {
	CreateLinkedStore(ctx context.Context, store MetadataStore) (linking.LinkSystem, error)
}

type BadgerLinkedStoreConfig struct{}

func (b BadgerLinkedStoreConfig) CreateLinkedStore(ctx context.Context, store MetadataStore) (linking.LinkSystem, error) {
	dsa := &dsadapter.Adapter{
		Wrapped: store,

		EscapingFunc: func(s string) string {
			return "_cas/" + hex.EncodeToString([]byte(s))
		},
	}

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetReadStorage(dsa)
	lsys.SetWriteStorage(dsa)
	lsys.TrustedStorage = true

	return lsys, nil
}
