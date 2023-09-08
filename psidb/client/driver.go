package client

import (
	"context"
	"sync"

	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/ipld/go-ipld-prime/linking"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/psidb/apis/rt/v1"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/adapters/remotesb"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Driver struct {
	client *rtv1.Client

	sp    inject.ServiceLocator
	types psi.TypeRegistry
	root  psi.Path

	metadataStore *dssync.MutexDatastore
	linkSystem    linking.LinkSystem
	virtualGraph  *graphfs.VirtualGraph

	superBlockMapMutex sync.RWMutex
	superBlockMap      map[string]graphfs.SuperBlock
}

func NewDriver(
	client *rtv1.Client,
	sp inject.ServiceLocator,
	types psi.TypeRegistry,
	root psi.Path,
) (*Driver, error) {
	d := &Driver{
		sp:     sp,
		types:  types,
		client: client,
		root:   root,

		metadataStore: dssync.MutexWrap(datastore.NewMapDatastore()),
	}

	ckpt := coreapi.NewInMemoryCheckpoint()
	vg, err := graphfs.NewVirtualGraph(&d.linkSystem, d.provideSuperBlock, nil, ckpt, d.metadataStore)

	if err != nil {
		return nil, err
	}

	d.virtualGraph = vg

	return d, nil
}

func (d *Driver) Client() *rtv1.Client { return d.client }

func (d *Driver) BeginTransaction(ctx context.Context, options ...coreapi.TransactionOption) (coreapi.Transaction, error) {
	var opts coreapi.TransactionOptions

	opts.Apply(options...)

	lg, err := online.NewLiveGraph(ctx, d.root, d.virtualGraph, &d.linkSystem, d.types, d.sp)

	if err != nil {
		return nil, err
	}

	tx := &FramedTransaction{
		graph: lg,
	}

	return tx, nil
}

func (d *Driver) LookupNode(ctx context.Context, req *rtv1.LookupNodeRequest) (*rtv1.LookupNodeResponse, error) {
	return rtv1.MessageTypeLookupNode.MakeCall(ctx, d.client, req)
}

func (d *Driver) ReadNode(ctx context.Context, req *rtv1.ReadNodeRequest) (*rtv1.ReadNodeResponse, error) {
	return rtv1.MessageTypeReadNode.MakeCall(ctx, d.client, req)
}

func (d *Driver) ReadEdge(ctx context.Context, req *rtv1.ReadEdgeRequest) (*rtv1.ReadEdgeResponse, error) {
	return rtv1.MessageTypeReadEdge.MakeCall(ctx, d.client, req)
}

func (d *Driver) ReadEdges(ctx context.Context, req *rtv1.ReadEdgesRequest) (*rtv1.ReadEdgesResponse, error) {
	return rtv1.MessageTypeReadEdges.MakeCall(ctx, d.client, req)
}

func (d *Driver) ReadEdgesStream(ctx context.Context, req *rtv1.ReadEdgesRequest) (iterators.Iterator[*rtv1.ReadEdgesResponse], error) {
	ch, err := rtv1.MessageTypeReadEdges.MakeCallStreamed(ctx, d.client, req, 16)

	if err != nil {
		return nil, err
	}

	it := iterators.FlatMap(iterators.FromChannel(ch), func(t rtv1.RpcResponse) iterators.Iterator[*rtv1.ReadEdgesResponse] {
		return t.Result.(iterators.Iterator[*rtv1.ReadEdgesResponse])
	})

	return it, nil
}

func (d *Driver) PushFrame(ctx context.Context, req *rtv1.PushFrameRequest) (*rtv1.PushFrameResponse, error) {
	return rtv1.MessageTypePushFrame.MakeCall(ctx, d.client, req)
}

func (d *Driver) Close() error {
	return nil
}

func (d *Driver) provideSuperBlock(ctx context.Context, uuid string) (graphfs.SuperBlock, error) {
	d.superBlockMapMutex.RLock()
	if sb := d.superBlockMap[uuid]; sb != nil {
		d.superBlockMapMutex.RUnlock()
		return sb, nil
	}
	d.superBlockMapMutex.RUnlock()

	d.superBlockMapMutex.Lock()
	defer d.superBlockMapMutex.Unlock()

	if sb := d.superBlockMap[uuid]; sb != nil {
		return sb, nil
	}

	sb := remotesb.NewRemoteSuperBlock(d, uuid)

	d.superBlockMap[uuid] = sb

	return sb, nil
}
