package core

import (
	"context"
	"path"
	"sync"

	"github.com/ipld/go-ipld-prime/linking"
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"github.com/pkg/errors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/adapters/psidsadapter"
	"github.com/greenboxal/agibootstrap/psidb/db/journal"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Core struct {
	mu sync.Mutex

	logger *otelzap.SugaredLogger
	tracer trace.Tracer

	srm    *inject.ServiceRegistrationManager
	config *coreapi.Config

	serviceProvider inject.ServiceProvider
	rootSession     coreapi.Session

	proc goprocess.Process

	closing bool
	closed  bool
	ready   bool

	readyCh chan struct{}
	closeCh chan struct{}
}

func NewCore(
	lc fx.Lifecycle,
	cfg *coreapi.Config,
	srm *inject.ServiceRegistrationManager,
) (*Core, error) {
	core := &Core{
		logger: logging.GetLogger("core"),
		tracer: otel.Tracer("core"),

		srm:    srm,
		config: cfg,

		readyCh: make(chan struct{}),
		closeCh: make(chan struct{}),
	}

	lc.Append(fx.Hook{
		OnStart: core.Start,
		OnStop:  core.Stop,
	})

	return core, nil
}

func (c *Core) Ready() <-chan struct{} { return c.readyCh }
func (c *Core) IsReady() bool          { return c.ready }

func (c *Core) Config() *coreapi.Config                 { return c.config }
func (c *Core) Journal() coreapi.Journal                { return c.rootSession.Journal() }
func (c *Core) VirtualGraph() coreapi.VirtualGraph      { return c.rootSession.VirtualGraph() }
func (c *Core) LinkSystem() *linking.LinkSystem         { return c.rootSession.LinkSystem() }
func (c *Core) MetadataStore() coreapi.DataStore        { return c.rootSession.MetadataStore() }
func (c *Core) ServiceProvider() inject.ServiceProvider { return c.serviceProvider }

func (c *Core) CreateConfirmationTracker(ctx context.Context, name string) (coreapi.ConfirmationTracker, error) {
	if err := c.waitReady(ctx); err != nil {
		return nil, err
	}

	ct, err := NewConfirmationTracker(ctx, c.MetadataStore(), name)

	if err != nil {
		return nil, err
	}

	return ct, nil
}

func (c *Core) CreateReplicationSlot(ctx context.Context, options coreapi.ReplicationSlotOptions) (coreapi.ReplicationSlot, error) {
	if err := c.waitReady(ctx); err != nil {
		return nil, err
	}

	return c.VirtualGraph().CreateReplicationSlot(ctx, options)
}

func (c *Core) BeginTransaction(ctx context.Context, options ...coreapi.TransactionOption) (coreapi.Transaction, error) {
	sess := coreapi.GetSession(ctx)

	if sess != nil {
		return sess.BeginTransaction(ctx, options...)
	}

	return c.rootSession.BeginTransaction(ctx, options...)
}

func (c *Core) RunTransaction(ctx context.Context, fn coreapi.TransactionFunc, options ...coreapi.TransactionOption) (err error) {
	return coreapi.RunTransaction(ctx, c, fn, options...)
}

func (c *Core) Start(ctx context.Context) error {
	c.serviceProvider = c.srm.Global.Build()

	c.proc = goprocess.Go(c.run)

	return c.waitReady(ctx)
}

func (c *Core) Stop(ctx context.Context) error {
	c.mu.Lock()

	if c.closing {
		c.mu.Unlock()
		return nil
	}

	close(c.closeCh)
	c.closing = true
	c.mu.Unlock()

	return c.proc.Close()
}

func (c *Core) buildSessionConfig() coreapi.SessionConfig {
	return coreapi.SessionConfig{
		SessionID:       "",
		ParentSessionID: "",
		Persistent:      true,

		Root: psi.PathFromElements(c.config.RootUUID, false),

		MetadataStore: c.buildMetadataStoreConfig(),
		GraphStore:    c.buildGraphStoreConfig(),
		Checkpoint:    c.buildCheckpointConfig(),
		Journal:       c.buildJournalConfig(),

		MountPoints: []coreapi.MountDefinition{
			c.buildDataMountPoint(c.config.RootUUID),
		},
	}
}

func (c *Core) buildCheckpointConfig() coreapi.CheckpointConfig {
	return coreapi.FileCheckpointConfig{
		Path: path.Join(c.config.DataDir, "psidb.ckpt"),
	}
}

func (c *Core) buildJournalConfig() coreapi.JournalConfig {
	return journal.FileJournalConfig{
		Path: path.Join(c.config.DataDir, "journal"),
	}
}

func (c *Core) buildGraphStoreConfig() coreapi.DataStoreConfig {
	return coreapi.BadgerDataStoreConfig{
		Path: path.Join(c.config.DataDir, "data"),
	}
}

func (c *Core) buildMetadataStoreConfig() coreapi.DataStoreConfig {
	return coreapi.BadgerDataStoreConfig{
		Path: path.Join(c.config.DataDir, "metadata"),
	}
}

func (c *Core) buildDataMountPoint(name string) coreapi.MountDefinition {
	return coreapi.MountDefinition{
		Name: name,
		Path: psi.PathFromElements(c.config.RootUUID, false),
		Target: psidsadapter.BadgerSuperBlockConfig{
			MetadataStoreConfig: coreapi.BadgerDataStoreConfig{
				Path: path.Join(c.config.DataDir, "blocks", name),
			},
		},
	}
}

func (c *Core) run(proc goprocess.Process) {
	proc.SetTeardown(c.teardown)

	ctx := goprocessctx.OnClosingContext(proc)

	sessionManager := inject.Inject[coreapi.SessionManager](c.serviceProvider)
	sessionConfig := c.buildSessionConfig()
	c.rootSession = sessionManager.GetOrCreateSession(sessionConfig)

	err := c.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
		root, err := tx.Resolve(ctx, psi.PathFromElements(c.config.RootUUID, false))

		if err != nil && !errors.Is(err, psi.ErrNodeNotFound) {
			return err
		}

		if root == nil {
			r := &stdlib.RootNode{NodeUUID: c.config.RootUUID}
			r.Init(r)

			tx.Add(r)

			if err := r.Update(ctx); err != nil {
				return err
			}

			root = r
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	c.ready = true
	close(c.readyCh)

	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			panic(err)
		}

	case _, _ = <-c.closeCh:
		return
	}
}

func (c *Core) waitReady(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case _, _ = <-c.readyCh:
		if c.closed {
			return errors.New("core closed")
		}

		return nil
	}
}

func (c *Core) teardown() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Error("Initiating shutdown")

	c.closed = true

	return nil
}
