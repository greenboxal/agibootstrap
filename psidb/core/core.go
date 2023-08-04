package core

import (
	"context"
	"encoding/hex"

	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/dsadapter"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
	graphfs "github.com/greenboxal/agibootstrap/psidb/db/graphfs"
	"github.com/greenboxal/agibootstrap/psidb/db/indexing"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
)

type Core struct {
	logger *zap.SugaredLogger

	cfg *coreapi.Config

	ds   coreapi.DataStore
	lsys linking.LinkSystem

	journal      *graphfs.Journal
	checkpoint   graphfs.Checkpoint
	virtualGraph *graphfs.VirtualGraph

	sp inject.ServiceProvider
}

func NewCore(
	lc fx.Lifecycle,
	sp inject.ServiceProvider,
	cfg *coreapi.Config,
	ds coreapi.DataStore,
	journal *graphfs.Journal,
	checkpoint graphfs.Checkpoint,
	blockManager *BlockManager,
) (*Core, error) {
	dsa := &dsadapter.Adapter{
		Wrapped: ds,

		EscapingFunc: func(s string) string {
			return "_cas/" + hex.EncodeToString([]byte(s))
		},
	}

	core := &Core{
		logger: logging.GetLogger("core"),

		cfg:        cfg,
		ds:         ds,
		sp:         sp,
		journal:    journal,
		checkpoint: checkpoint,
	}

	core.lsys = cidlink.DefaultLinkSystem()
	core.lsys.SetReadStorage(dsa)
	core.lsys.SetWriteStorage(dsa)
	core.lsys.TrustedStorage = true

	virtualGraph, err := graphfs.NewVirtualGraph(
		&core.lsys,
		blockManager.Resolve,
		journal,
		checkpoint,
		nil,
	)

	if err != nil {
		return nil, err
	}

	core.virtualGraph = virtualGraph

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return core.Start(ctx)
		},

		OnStop: func(ctx context.Context) error {
			return core.Stop(ctx)
		},
	})

	return core, nil
}

func (c *Core) Config() *coreapi.Config                 { return c.cfg }
func (c *Core) DataStore() coreapi.DataStore            { return c.ds }
func (c *Core) Journal() *graphfs.Journal               { return c.journal }
func (c *Core) Checkpoint() graphfs.Checkpoint          { return c.checkpoint }
func (c *Core) LinkSystem() *linking.LinkSystem         { return &c.lsys }
func (c *Core) VirtualGraph() *graphfs.VirtualGraph     { return c.virtualGraph }
func (c *Core) ServiceProvider() inject.ServiceProvider { return c.sp }

func (c *Core) BeginTransaction(ctx context.Context, options ...coreapi.TransactionOption) (coreapi.Transaction, error) {
	var opts coreapi.TransactionOptions

	opts.Apply(options...)

	tx := &transaction{
		core: c,
		opts: opts,
	}

	lg, err := online.NewLiveGraph(ctx, &c.lsys, c.virtualGraph, tx)

	if err != nil {
		return nil, err
	}

	tx.lg = lg

	return tx, nil
}

func (c *Core) RunTransaction(ctx context.Context, fn coreapi.TransactionFunc, options ...coreapi.TransactionOption) (err error) {
	return coreapi.RunTransaction(ctx, c, fn, options...)
}

func (c *Core) Start(ctx context.Context) error {
	if err := c.virtualGraph.Recover(ctx); err != nil {
		return err
	}

	return c.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
		root, err := tx.Resolve(ctx, psi.PathFromElements(c.cfg.RootUUID, false))

		if err != nil && err != psi.ErrNodeNotFound {
			return err
		}

		if root == nil {
			r := &stdlib.RootNode{NodeUUID: c.cfg.RootUUID}
			r.Init(r)

			tx.Add(r)

			scp := indexing.NewScope()
			scp.SetParent(r)

			indexing.SetNodeScope(r, scp)

			if err := r.Update(ctx); err != nil {
				return err
			}

			root = r
		}

		return nil
	})
}

func (c *Core) Stop(ctx context.Context) error {
	return c.virtualGraph.Close(ctx)
}
