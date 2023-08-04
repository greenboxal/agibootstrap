package network

import (
	"context"
	"errors"
	"fmt"
	"sync"

	ds "github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p-kad-dht/providers"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	routing2 "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"go.uber.org/zap"

	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/psidb/config"
)

type IpfsRouting interface {
	routing.ContentRouting
	routing.PeerRouting
	routing.ValueStore

	Bootstrap(context.Context) error
}

type Network struct {
	logger *zap.SugaredLogger

	cfg *config.Config
	lrm config.LocalResourceManager
	cm  *connmgr.BasicConnMgr

	host          host.Host
	mdns          mdns.Service
	disco         *routing2.RoutingDiscovery
	dht           *dual.DHT
	peerRouter    routing.PeerRouting
	providerStore *providers.ProviderManager
}

func NewNetwork(
	lc fx.Lifecycle,
	cfg *config.Config,
	lrm config.LocalResourceManager,
	cm *connmgr.BasicConnMgr,
) (*Network, error) {
	hf := &Network{
		lrm: lrm,
		cm:  cm,
	}

	hf.logger = logging.GetLogger("network/p2p")

	if err := hf.initializeHost(context.Background()); err != nil {
		return nil, err
	}

	if err := hf.initializeMdns(); err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return hf.Shutdown(ctx)
		},
	})

	return hf, nil
}

func (n *Network) initializeHost(ctx context.Context) error {
	h, err := libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(n.cfg.Identity),
		// Multiple listen addresses
		libp2p.NoListenAddrs,
		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),
		// support any other default transports (TCP)
		libp2p.DefaultPrivateTransports,
		// Let's prevent our peer from having too many
		// connections by attaching a connection manager.
		libp2p.ConnectionManager(n.cm),
		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),
		// Let this host use the DHT to find other hosts
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			return n.initializeDht(ctx, h)
		}),
		// If you want to help other peers to figure out if they are behind
		// NATs, you can launch the server-side of AutoNAT too (AutoRelay
		// already runs the client)
		//
		// This service is highly rate-limited and should not cause any
		// performance issues.
		libp2p.EnableNATService(),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelay(),
		libp2p.EnableRelayService(),
		//libp2p.EnableAutoRelayWithPeerSource(n.findBootstrapPeers),
	)

	if err != nil {
		return err
	}

	if h.ID() != n.cfg.PeerID {
		return errors.New("local peer key mismatch")
	}

	n.host = h

	return nil
}

func (n *Network) initializeDht(ctx context.Context, h host.Host) (routing.PeerRouting, error) {
	dataStore := dssync.MutexWrap(ds.NewMapDatastore())
	providerStore, err := providers.NewProviderManager(ctx, h.ID(), h.Peerstore(), dataStore)

	if err != nil {
		return nil, fmt.Errorf("initializing default provider manager (%v)", err)
	}

	if n.peerRouter != nil {
		return n.peerRouter, nil
	}

	dhtOptions := []dual.Option{
		dual.DHTOption(
			dht.BootstrapPeersFunc(n.getDhtBootstrapPeers),
			dht.ProviderStore(providerStore),
		),
	}

	d, err := dual.New(ctx, h, dhtOptions...)

	if err != nil {
		return nil, err
	}

	n.dht = d
	n.providerStore = providerStore
	n.peerRouter = n.dht

	return n.peerRouter, nil
}

func (n *Network) initializeMdns() error {
	n.mdns = mdns.NewMdnsService(n.host, "npnsd", n)

	n.disco = routing2.NewRoutingDiscovery(n.dht)

	return nil
}

func (n *Network) HandlePeerFound(info peer.AddrInfo) {
	n.host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.TempAddrTTL)
}

func (n *Network) Run(proc goprocess.Process) {
	ctx := goprocessctx.OnClosingContext(proc)

	addrs := n.lrm.NetworkListenAddresses()

	if err := n.host.Network().Listen(addrs...); err != nil {
		panic(err)
	}

	if err := n.dht.Bootstrap(ctx); err != nil {
		panic(err)
	}

	if err := n.mdns.Start(); err != nil {
		panic(err)
	}

	proc.Go(func(proc goprocess.Process) {
		var wg sync.WaitGroup

		for _, info := range n.getDhtBootstrapPeers() {
			wg.Add(1)

			go func(info peer.AddrInfo) {
				defer wg.Done()

				if err := n.host.Connect(ctx, info); err != nil {
					n.logger.Warn(err)
				}
			}(info)
		}
	})
}

func (n *Network) Shutdown(ctx context.Context) error {
	if n.dht != nil {
		if err := n.dht.Close(); err != nil {
			return err
		}

		n.dht = nil
	}

	if n.host != nil {
		if err := n.host.Close(); err != nil {
			return err
		}

		n.host = nil
	}

	return nil
}

func (n *Network) getDhtBootstrapPeers() []peer.AddrInfo {
	return nil
}
