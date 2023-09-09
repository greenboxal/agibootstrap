package network

import (
	"context"
	"path"

	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
	"go.uber.org/fx"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

func NewPeerStore(
	lc fx.Lifecycle,
	core coreapi.Core,
) (peerstore.Peerstore, error) {
	ctx := context.Background()
	opts := pstoreds.DefaultOpts()

	ds, err := coreapi.BadgerDataStoreConfig{
		Path: path.Join(core.Config().DataDir, "peerstore"),
	}.CreateDataStore(ctx)

	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return ds.Close()
		},
	})

	ps, err := pstoreds.NewPeerstore(ctx, ds, opts)

	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return ps.Close()
		},
	})

	return ps, nil
}
