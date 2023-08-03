package graphfs

import (
	"context"
	"sync"

	"github.com/ipld/go-ipld-prime/linking"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type SuperBlockProvider func(ctx context.Context, uuid string) (SuperBlock, error)

type VirtualGraph struct {
	lsys     *linking.LinkSystem
	spb      SuperBlockProvider
	listener VirtualGraphListener

	mu          sync.RWMutex
	txm         *TransactionManager
	superBlocks map[string]SuperBlock
}

type VirtualGraphListener interface {
	OnCommitTransaction(ctx context.Context, tx *Transaction) error
}

func NewVirtualGraph(
	lsys *linking.LinkSystem,
	spb SuperBlockProvider,
	journal *Journal,
	checkpoint Checkpoint,
	listener VirtualGraphListener,
) (*VirtualGraph, error) {
	vg := &VirtualGraph{
		lsys:     lsys,
		spb:      spb,
		listener: listener,

		superBlocks: map[string]SuperBlock{},
	}

	vg.txm = NewTransactionManager(vg, journal, checkpoint)

	return vg, nil
}

func (vg *VirtualGraph) Recover(ctx context.Context) error {
	return vg.txm.Recover(ctx)
}

func (vg *VirtualGraph) BeginTransaction(ctx context.Context) (*Transaction, error) {
	return vg.txm.BeginTransaction()
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
	var opts OpenNodeOptions

	opts.Transaction = GetTransaction(ctx)
	opts.Apply(options...)

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

	if sb == nil {
		return nil, errors.Errorf("no such superblock: %s", path.Root())
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

func (vg *VirtualGraph) applyTransaction(ctx context.Context, tx *Transaction) error {
	hasBegun := false
	hasFinished := false

	nodeByPath := map[string]NodeHandle{}
	nodeByHandle := map[int64]NodeHandle{}

	defer func() {
		for _, nh := range nodeByPath {
			if err := nh.Close(); err != nil {
				vg.txm.logger.Error(err)
			}
		}
	}()

	getHandle := func(entry *JournalEntry) NodeHandle {
		if entry.Path == nil {
			return nodeByHandle[entry.Inode]
		}

		str := entry.Path.String()

		if nh := nodeByPath[str]; nh != nil {
			return nh
		}

		nh, err := vg.Open(ctx, *entry.Path, WithOpenNodeCreateIfMissing(), WithOpenNodeForceInode(entry.Inode))

		if err != nil {
			panic(err)
		}

		nodeByPath[str] = nh
		nodeByHandle[nh.Inode().ID()] = nh

		return nh
	}

	for _, entry := range tx.log {
		if hasFinished {
			return errors.New("invalid transaction log")
		}

		if entry.Op != JournalOpBegin && !hasBegun {
			return errors.New("invalid transaction log")
		}

		switch entry.Op {
		case JournalOpBegin:
			hasBegun = true

		case JournalOpCommit:
			hasFinished = true

		case JournalOpRollback:
			hasFinished = true

		case JournalOpWrite:
			nh := getHandle(entry)

			if nh == nil {
				return errors.New("invalid transaction log")
			}

			if err := nh.Write(ctx, entry.Node); err != nil {
				return err
			}

		case JournalOpSetEdge:
			nh := getHandle(entry)

			if nh == nil {
				return errors.New("invalid transaction log")
			}

			if err := nh.SetEdge(ctx, entry.Edge); err != nil {
				return err
			}

		case JournalOpRemoveEdge:
			nh := getHandle(entry)

			if nh == nil {
				return errors.New("invalid transaction log")
			}

			if err := nh.RemoveEdge(ctx, entry.Edge.Key); err != nil {
				return err
			}
		}
	}

	if !hasBegun || !hasFinished {
		return errors.New("invalid transaction log")
	}

	return vg.listener.OnCommitTransaction(ctx, tx)
}
