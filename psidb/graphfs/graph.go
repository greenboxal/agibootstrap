package graphfs

import (
	"context"
	"sync"

	"github.com/ipld/go-ipld-prime"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type SuperBlockProvider func(ctx context.Context, uuid string) (SuperBlock, error)

type VirtualGraph struct {
	spb SuperBlockProvider

	mu          sync.RWMutex
	superBlocks map[string]SuperBlock
}

func NewVirtualGraph(spb SuperBlockProvider) *VirtualGraph {
	return &VirtualGraph{
		spb:         spb,
		superBlocks: map[string]SuperBlock{},
	}
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

func (vg *VirtualGraph) Write(ctx context.Context, path psi.Path, node *SerializedNode) (ipld.Link, error) {
	nh, err := vg.Open(ctx, path)

	if err != nil {
		return nil, err
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
