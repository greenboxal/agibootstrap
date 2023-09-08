package session

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/ipld/go-ipld-prime/linking"
	"github.com/jbenet/goprocess"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
	vm "github.com/greenboxal/agibootstrap/psidb/core/vm"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/typing"
)

type SessionStatus int

const (
	SessionStateNew SessionStatus = iota
	SessionStateInitialized
	SessionStateActive
	SessionStateClosing
	SessionStateClosed
)

type SessionState struct {
	Config        coreapi.SessionConfig `json:"config,omitempty"`
	Status        SessionStatus         `json:"status,omitempty"`
	LastKeepAlive time.Time             `json:"lastKeepAlive,omitempty"`
}

type Session struct {
	mu sync.RWMutex

	logger *otelzap.SugaredLogger

	manager *Manager
	core    coreapi.Core

	config   coreapi.SessionConfig
	parent   *Session
	children []*Session

	proc   goprocess.Process
	state  SessionState
	status SessionStatus

	readyCh chan struct{}
	stopCh  chan struct{}

	lastKeepAlive   time.Time
	serviceProvider inject.ServiceProvider

	metadataStore coreapi.MetadataStore
	journal       coreapi.Journal
	checkpoint    coreapi.Checkpoint
	blockManager  *MountManager
	virtualGraph  *graphfs.VirtualGraph

	*ClientBusConnection
}

func NewSession(manager *Manager, parent *Session, config coreapi.SessionConfig) *Session {
	parentSp := manager.core.ServiceProvider()

	if parent != nil {
		parentSp = parent.serviceProvider
	}

	serviceProvider := manager.srm.Session.Build(inject.WithParentServiceProvider(parentSp))

	sess := &Session{
		logger: logging.GetLogger("session"),

		config:  config,
		parent:  parent,
		manager: manager,
		core:    manager.core,

		status:  SessionStateNew,
		readyCh: make(chan struct{}),
		stopCh:  make(chan struct{}),

		serviceProvider: serviceProvider,
		blockManager:    NewBlockManager(),

		ClientBusConnection: NewClientBusConnection(config.SessionID, 16, 16),
	}

	inject.RegisterInstance(sess.serviceProvider, sess)

	sess.proc = goprocess.Go(sess.Run)

	return sess
}

func (sess *Session) UUID() string             { return sess.config.SessionID }
func (sess *Session) KeepAlive()               { sess.lastKeepAlive = time.Now() }
func (sess *Session) LastKeepAlive() time.Time { return sess.lastKeepAlive }

func (sess *Session) Journal() coreapi.Journal                { return sess.journal }
func (sess *Session) Checkpoint() coreapi.Checkpoint          { return sess.checkpoint }
func (sess *Session) MetadataStore() coreapi.MetadataStore    { return sess.metadataStore }
func (sess *Session) LinkSystem() *linking.LinkSystem         { return sess.metadataStore.LinkSystem() }
func (sess *Session) VirtualGraph() coreapi.VirtualGraph      { return sess.virtualGraph }
func (sess *Session) ServiceProvider() inject.ServiceProvider { return sess.serviceProvider }
func (sess *Session) ServiceLocator() inject.ServiceLocator   { return sess.serviceProvider }

func (sess *Session) Closed() <-chan struct{}  { return sess.proc.Closed() }
func (sess *Session) Closing() <-chan struct{} { return sess.stopCh }
func (sess *Session) Ready() <-chan struct{}   { return sess.readyCh }
func (sess *Session) Err() error               { return sess.proc.Err() }

func (sess *Session) Fork(config coreapi.SessionConfig) coreapi.Session {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	for _, child := range sess.children {
		if child.UUID() == config.SessionID {
			// TODO: Assert that the config is the same
			return child
		}
	}

	child := NewSession(sess.manager, sess, sess.config.Extend(config))

	sess.children = append(sess.children, child)

	return child
}

func (sess *Session) BeginTransaction(ctx context.Context, options ...coreapi.TransactionOption) (coreapi.Transaction, error) {
	var err error
	var opts coreapi.TransactionOptions

	opts.Root = sess.config.Root
	opts.Apply(options...)

	if opts.Root.IsEmpty() {
		opts.Root = sess.config.Root
	} else if opts.Root.IsRelative() {
		opts.Root = sess.config.Root.Join(opts.Root)
	}

	tx := &transaction{
		session: sess,
		core:    sess.core,
		opts:    opts,
	}

	tx.sp = inject.NewServiceProvider(
		inject.WithParentServiceProvider(sess.serviceProvider),
		inject.WithServiceRegistry(&sess.manager.srm.Transaction),
	)

	ctx = coreapi.WithSession(ctx, sess)
	ctx = coreapi.WithTransaction(ctx, tx)

	if err := sess.waitReady(ctx); err != nil {
		return nil, err
	}

	typingManager := inject.Inject[*typing.Manager](tx.sp)
	typeRegistry := vm.NewTypeRegistry(typingManager)

	tx.lg, err = online.NewLiveGraph(
		ctx,
		opts.Root,
		sess.virtualGraph,
		sess.metadataStore.LinkSystem(),
		typeRegistry,
		tx.sp,
	)

	if err != nil {
		return nil, err
	}

	inject.Register[*vm.Context](tx.sp, func(iso *vm.Isolate) (*vm.Context, error) {
		vmctx := vm.NewContext(ctx, iso, tx.sp)

		vm.MustSet(vmctx.Global(), "_psidb_tx", vmctx.MustWrapValue(reflect.ValueOf(coreapi.Transaction(tx))))
		vm.MustSet(vmctx.Global(), "_psidb_sp", vmctx.MustWrapValue(reflect.ValueOf(tx.ServiceLocator())))

		return vmctx, nil
	})

	inject.RegisterInstance[psi.TypeRegistry](tx.sp, typeRegistry)
	inject.RegisterInstance[psi.Graph](tx.sp, tx.lg)

	return tx, nil
}

func (sess *Session) RunTransaction(ctx context.Context, fn coreapi.TransactionFunc, options ...coreapi.TransactionOption) (err error) {
	return coreapi.RunTransaction(ctx, sess, fn, options...)
}

func (sess *Session) initialize(ctx context.Context) error {
	var err error

	sess.mu.Lock()
	defer sess.mu.Unlock()

	if sess.status != SessionStateNew {
		return nil
	}

	if sess.parent != nil {
		for _, mount := range sess.parent.blockManager.mountTab {
			sess.blockManager.mountTab[mount.Name] = mount
		}
	}

	for _, md := range sess.config.MountPoints {
		sess.blockManager.RegisterMountDefinition(md)
	}

	if sess.config.MetadataStore != nil {
		sess.metadataStore, err = sess.config.MetadataStore.CreateMetadataStore(ctx)

		if err != nil {
			return err
		}

		sess.serviceProvider.AppendShutdownHook(func(ctx context.Context) error {
			return sess.metadataStore.Close()
		})
	} else if sess.parent != nil {
		sess.metadataStore = sess.parent.metadataStore
	} else {
		return errors.New("no metadata store configured")
	}

	if sess.config.Journal != nil {
		sess.journal, err = sess.config.Journal.CreateJournal(ctx)

		if err != nil {
			return err
		}

		sess.serviceProvider.AppendShutdownHook(func(ctx context.Context) error {
			return sess.journal.Close()
		})
	} else if sess.parent != nil {
		sess.journal = sess.parent.journal
	} else {
		return errors.New("no journal configured")
	}

	if sess.config.Checkpoint != nil {
		sess.checkpoint, err = sess.config.Checkpoint.CreateCheckpoint(ctx)

		if err != nil {
			return err
		}

		sess.serviceProvider.AppendShutdownHook(func(ctx context.Context) error {
			return sess.checkpoint.Close()
		})
	} else if sess.parent != nil {
		sess.checkpoint = sess.parent.checkpoint
	} else {
		return errors.New("no checkpoint configured")
	}

	sess.virtualGraph, err = graphfs.NewVirtualGraph(
		sess.metadataStore.LinkSystem(),
		sess.blockManager.Resolve,
		sess.journal,
		sess.checkpoint,
		sess.metadataStore,
	)

	if err != nil {
		return err
	}

	sess.serviceProvider.AppendShutdownHook(func(ctx context.Context) error {
		return sess.virtualGraph.Close(ctx)
	})

	if err := sess.initializeServices(ctx); err != nil {
		return err
	}

	sess.status = SessionStateInitialized

	return nil
}

func (sess *Session) initializeServices(ctx context.Context) error {
	inject.Register[*vm.Isolate](sess.serviceProvider, func(ctx inject.ResolutionContext, vms *vm.Supervisor) (*vm.Isolate, error) {
		return vms.NewIsolate(), nil
	})

	return nil
}

func (sess *Session) waitReady(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()

	case _, _ = <-sess.readyCh:

		if sess.status != SessionStateActive {
			return errors.New("session not active")
		}

		return nil
	}
}
