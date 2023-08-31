package client

import (
	"context"
	"sync"

	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/ipld/go-ipld-prime/linking"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
	"github.com/greenboxal/agibootstrap/psidb/db/providers/remotesb"
)

type Driver struct {
	client *Client

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
	client *Client,
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

func (d *Driver) Client() *Client { return d.client }

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

func (d *Driver) LookupNode(ctx context.Context, req *LookupNodeRequest) (*LookupNodeResponse, error) {
	return MessageTypeLookupNode.MakeCall(ctx, d.client, req)
}

func (d *Driver) ReadNode(ctx context.Context, req *ReadNodeRequest) (*ReadNodeResponse, error) {
	return MessageTypeReadNode.MakeCall(ctx, d.client, req)
}

func (d *Driver) ReadEdge(ctx context.Context, req *ReadEdgeRequest) (*ReadEdgeResponse, error) {
	return MessageTypeReadEdge.MakeCall(ctx, d.client, req)
}

func (d *Driver) ReadEdges(ctx context.Context, req *ReadEdgesRequest) (*ReadEdgesResponse, error) {
	return MessageTypeReadEdges.MakeCall(ctx, d.client, req)
}

func (d *Driver) ReadEdgesStream(ctx context.Context, req *ReadEdgesRequest) (iterators.Iterator[*ReadEdgesResponse], error) {
	ch, err := MessageTypeReadEdges.MakeCallStreamed(ctx, d.client, req, 16)

	if err != nil {
		return nil, err
	}

	it := iterators.FlatMap(iterators.FromChannel(ch), func(t RpcResponse) iterators.Iterator[*ReadEdgesResponse] {
		return t.Result.(iterators.Iterator[*ReadEdgesResponse])
	})

	return it, nil
}

func (d *Driver) PushFrame(ctx context.Context, req *PushFrameRequest) (*PushFrameResponse, error) {
	return MessageTypePushFrame.MakeCall(ctx, d.client, req)
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
