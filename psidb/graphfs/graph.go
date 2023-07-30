package graphfs

import (
	"context"
	"sync"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type SuperBlockProvider func(ctx context.Context, uuid string) (SuperBlock, error)

type VirtualGraph struct {
	spb SuperBlockProvider

	mu          sync.RWMutex
	txm         *TransactionManager
	superBlocks map[string]SuperBlock
}

func NewVirtualGraph(spb SuperBlockProvider, journal *Journal, checkpoint Checkpoint) *VirtualGraph {
	vg := &VirtualGraph{
		spb: spb,

		superBlocks: map[string]SuperBlock{},
	}

	vg.txm = NewTransactionManager(vg, journal, checkpoint)

	return vg
}

func (vg *VirtualGraph) GetSuperBlock(ctx context.Context, uuid string) (SuperBlock, error) {
	vg.mu.Lock()
	defer vg.mu.Unlock()

	if sb := vg.superBlocks[uuid]; sb != nil {
		return sb, nil
	}

	sb, err := vg.spb(ctx, uuid)

	if err != nil {
		return nil, err
	}

	if sb == nil {
		return nil, nil
	}

	vg.superBlocks[uuid] = sb

	return sb, nil
}

func (vg *VirtualGraph) Open(ctx context.Context, path psi.Path, options ...OpenNodeOption) (NodeHandle, error) {
	opts := NewOpenNodeOptions(options...)
	dentry, err := vg.Resolve(ctx, path)

	if err != nil {
		return nil, err
	}

	return dentry.sb.INodeOperations().Create(ctx, dentry, opts)
}

func (vg *VirtualGraph) Resolve(ctx context.Context, path psi.Path) (*CacheEntry, error) {
	sb, err := vg.GetSuperBlock(ctx, path.Root())

	if err != nil {
		return nil, err
	}

	root, err := sb.GetRoot(ctx)

	if err != nil {
		return nil, err
	}

	return Resolve(ctx, root, path)
}

func (vg *VirtualGraph) Read(ctx context.Context, path psi.Path) (*SerializedNode, error) {
	nh, err := vg.Open(ctx, path)

	if err != nil {
		return nil, err
	}

	defer nh.Close()

	return nh.Read(ctx)
}

func (vg *VirtualGraph) Write(ctx context.Context, path psi.Path, node *SerializedNode) error {
	nh, err := vg.Open(ctx, path)

	if err != nil {
		return err
	}

	defer nh.Close()

	return nh.Write(ctx, node)
}

func (vg *VirtualGraph) ReadEdge(ctx context.Context, path psi.Path) (*SerializedEdge, error) {
	nh, err := vg.Open(ctx, path.Parent())

	if err != nil {
		return nil, err
	}

	defer nh.Close()

	return nh.ReadEdge(ctx, path.Name())
}

func (vg *VirtualGraph) ReadEdges(ctx context.Context, path psi.Path) (iterators.Iterator[*SerializedEdge], error) {
	nh, err := vg.Open(ctx, path)

	if err != nil {
		return nil, err
	}

	defer nh.Close()

	return nh.ReadEdges(ctx)
}

func (vg *VirtualGraph) Close(ctx context.Context) error {
	if vg.txm != nil {
		if err := vg.txm.Close(ctx); err != nil {
			return err
		}
		vg.txm = nil
	}

	for _, sb := range vg.superBlocks {
		if err := sb.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}
