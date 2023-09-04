package graphfs

import (
	"context"
	"io/fs"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime/linking"
	"github.com/pkg/errors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
)

type SuperBlockProvider func(ctx context.Context, uuid string) (SuperBlock, error)

type VirtualGraph struct {
	mu sync.RWMutex

	logger *otelzap.SugaredLogger
	tracer trace.Tracer

	lsys *linking.LinkSystem
	spb  SuperBlockProvider

	transactionManager *TransactionManager
	replicationManager *replicationManager
	superBlocks        map[string]SuperBlock

	ds datastore.Batching
}

type VirtualGraphListener interface {
	OnCommitTransaction(ctx context.Context, tx *Transaction) error
}

func NewVirtualGraph(
	lsys *linking.LinkSystem,
	spb SuperBlockProvider,
	journal coreapi.Journal,
	checkpoint coreapi.Checkpoint,
	metadataStore datastore.Batching,
) (*VirtualGraph, error) {
	vg := &VirtualGraph{
		logger: logging.GetLogger("graphfs"),

		tracer: otel.Tracer("graphfs", trace.WithInstrumentationAttributes(
			semconv.ServiceName("psidb-graph"),
			semconv.DBSystemKey.String("psidb"),
		)),

		ds:   metadataStore,
		lsys: lsys,
		spb:  spb,

		superBlocks: map[string]SuperBlock{},
	}

	vg.transactionManager = NewTransactionManager(vg, journal, checkpoint)
	vg.replicationManager = newReplicationManager(vg)

	return vg, nil
}

func (vg *VirtualGraph) Recover(ctx context.Context) error {
	ctx, span := vg.tracer.Start(ctx, "VirtualGraph.Recover")
	span.SetAttributes(semconv.DBOperation("recover"))
	defer span.End()

	return vg.transactionManager.Recover(ctx)
}

func (vg *VirtualGraph) BeginTransaction(ctx context.Context) (*Transaction, error) {
	return vg.transactionManager.BeginTransaction(ctx)
}

func (vg *VirtualGraph) CreateReplicationSlot(ctx context.Context, options coreapi.ReplicationSlotOptions) (coreapi.ReplicationSlot, error) {
	return vg.replicationManager.CreateReplicationSlot(ctx, options)
}

func (vg *VirtualGraph) GetSuperBlock(ctx context.Context, uuid string) (SuperBlock, error) {
	ctx, span := vg.tracer.Start(ctx, "VirtualGraph.GetSuperBlock")
	span.SetAttributes(semconv.DBOperation("getSuperBlock"))
	defer span.End()

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

	ctx, span := vg.tracer.Start(ctx, "VirtualGraph.Open")
	span.SetAttributes(semconv.DBOperation("open"))
	defer span.End()

	if opts.Flags == 0 {
		opts.Flags = OpenNodeFlagsRead
	}

	opts.Apply(options...)

	dentry, err := vg.Resolve(ctx, path)

	if err != nil {
		return nil, err
	}

	if dentry.IsNegative() {
		if opts.Flags&OpenNodeFlagsCreate == 0 {
			return nil, psi.ErrNodeNotFound
		}

		err := dentry.sb.INodeOperations().Create(ctx, dentry, opts)

		if err != nil {
			return nil, err
		}
	} else if opts.Flags&OpenNodeFlagsAppend == 0 && opts.Flags&OpenNodeFlagsWrite != 0 {
		return nil, fs.ErrExist
	}

	return NewNodeHandle(ctx, dentry.Inode(), dentry, opts)
}

func (vg *VirtualGraph) Resolve(ctx context.Context, path psi.Path) (*CacheEntry, error) {
	ctx, span := vg.tracer.Start(ctx, "VirtualGraph.Resolve")
	span.SetAttributes(semconv.DBOperation("resolve"))
	defer span.End()

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

func (vg *VirtualGraph) Read(ctx context.Context, path psi.Path) (*coreapi.SerializedNode, error) {
	ctx, span := vg.tracer.Start(ctx, "VirtualGraph.Read")
	span.SetAttributes(semconv.DBOperation("read"))
	defer span.End()

	nh, err := vg.Open(ctx, path)

	if err != nil {
		return nil, err
	}

	defer nh.Close()

	return nh.Read(ctx)
}

func (vg *VirtualGraph) Write(ctx context.Context, path psi.Path, node *coreapi.SerializedNode) error {
	ctx, span := vg.tracer.Start(ctx, "VirtualGraph.Write")
	span.SetAttributes(semconv.DBOperation("write"))
	defer span.End()

	nh, err := vg.Open(ctx, path)

	if err != nil {
		return err
	}

	defer nh.Close()

	return nh.Write(ctx, node)
}

func (vg *VirtualGraph) ReadEdge(ctx context.Context, path psi.Path) (*coreapi.SerializedEdge, error) {
	ctx, span := vg.tracer.Start(ctx, "VirtualGraph.ReadEdge")
	span.SetAttributes(semconv.DBOperation("readEdge"))
	defer span.End()

	nh, err := vg.Open(ctx, path.Parent())

	if err != nil {
		return nil, err
	}

	defer nh.Close()

	return nh.ReadEdge(ctx, path.Name())
}

func (vg *VirtualGraph) ReadEdges(ctx context.Context, path psi.Path) (iterators.Iterator[*coreapi.SerializedEdge], error) {
	ctx, span := vg.tracer.Start(ctx, "VirtualGraph.ReadEdges")
	defer span.End()

	nh, err := vg.Open(ctx, path)

	if err != nil {
		return nil, err
	}

	defer nh.Close()

	return nh.ReadEdges(ctx)
}

func (vg *VirtualGraph) Close(ctx context.Context) error {
	if vg.replicationManager != nil {
		if err := vg.replicationManager.Close(ctx); err != nil {
			return err
		}

		vg.replicationManager = nil
	}

	if vg.transactionManager != nil {
		if err := vg.transactionManager.Close(ctx); err != nil {
			return err
		}
		vg.transactionManager = nil
	}

	for _, sb := range vg.superBlocks {
		if err := sb.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (vg *VirtualGraph) applyTransaction(ctx context.Context, tx *Transaction) error {
	ctx, span := vg.tracer.Start(ctx, "VirtualGraph.applyTransaction")
	span.SetAttributes(semconv.DBSystemKey.String("psidb"))
	defer span.End()

	hasBegun := false
	hasFinished := false

	superblocks := map[string]SuperBlock{}
	nodeByPath := map[string]NodeHandle{}
	nodeByHandle := map[int64]NodeHandle{}

	defer func() {
		for _, nh := range nodeByPath {
			if err := nh.Close(); err != nil {
				vg.transactionManager.logger.Error(err)
			}
		}
	}()

	getHandle := func(entry *coreapi.JournalEntry) NodeHandle {
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

		if entry.Op != coreapi.JournalOpBegin && !hasBegun {
			return errors.New("invalid transaction log")
		}

		switch entry.Op {
		case coreapi.JournalOpBegin:
			hasBegun = true

		case coreapi.JournalOpCommit:
			hasFinished = true

		case coreapi.JournalOpRollback:
			hasFinished = true

		case coreapi.JournalOpWrite:
			nh := getHandle(entry)

			if nh == nil {
				return errors.New("invalid transaction log")
			}

			if err := nh.Write(ctx, entry.Node); err != nil {
				return err
			}

			sb := nh.Inode().SuperBlock()
			superblocks[sb.UUID()] = sb

		case coreapi.JournalOpSetEdge:
			nh := getHandle(entry)

			if nh == nil {
				return errors.New("invalid transaction log")
			}

			if err := nh.SetEdge(ctx, entry.Edge); err != nil {
				return err
			}

			sb := nh.Inode().SuperBlock()
			superblocks[sb.UUID()] = sb

		case coreapi.JournalOpRemoveEdge:
			nh := getHandle(entry)

			if nh == nil {
				return errors.New("invalid transaction log")
			}

			if err := nh.RemoveEdge(ctx, entry.Edge.Key); err != nil {
				return err
			}

			sb := nh.Inode().SuperBlock()
			superblocks[sb.UUID()] = sb
		}
	}

	for _, sb := range superblocks {
		if err := sb.Flush(ctx); err != nil {
			return err
		}
	}

	return nil
}
