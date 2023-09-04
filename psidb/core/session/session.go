package session

import (
	"context"
	"sync"
	"time"

	`github.com/ipld/go-ipld-prime/linking`
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	`github.com/pkg/errors`
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	`golang.org/x/exp/slices`

	`github.com/greenboxal/agibootstrap/pkg/platform/inject`
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
	vm `github.com/greenboxal/agibootstrap/psidb/core/vm`
	`github.com/greenboxal/agibootstrap/psidb/db/graphfs`
	`github.com/greenboxal/agibootstrap/psidb/db/online`
	`github.com/greenboxal/agibootstrap/psidb/services/typing`
)

type SessionState int

const (
	SessionStateNew SessionState = iota
	SessionStateInitialized
	SessionStateActive
	SessionStateClosing
	SessionStateClosed
)

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
	stopCh chan struct{}

	lastKeepAlive time.Time

	serviceProvider inject.ServiceProvider

	lsys          linking.LinkSystem
	metadataStore coreapi.MetadataStore
	journal       coreapi.Journal
	checkpoint    coreapi.Checkpoint
	blockManager  *BlockManager
	virtualGraph  *graphfs.VirtualGraph

	*ClientBusConnection
}

func NewSession(manager *Manager, parent *Session, config coreapi.SessionConfig) *Session {
	parentSp := manager.core.ServiceProvider()

	if parent != nil {
		parentSp = parent.serviceProvider
	}

	sess := &Session{
		logger: logging.GetLogger("session"),

		config:  config,
		parent:  parent,
		manager: manager,
		core:    manager.core,

		state:  SessionStateNew,
		stopCh: make(chan struct{}),

		serviceProvider: inject.NewServiceProvider(inject.WithParentServiceProvider(parentSp)),
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

func (sess *Session) Run(proc goprocess.Process) {
	ctx := goprocessctx.OnClosingContext(proc)
	ctx = coreapi.WithSession(ctx, sess)

	defer func() {
		sess.mu.Lock()
		defer sess.mu.Unlock()

		if err := sess.teardown(ctx); err != nil {
			panic(err)
		}

		sess.manager.onSessionFinish(sess)
	}()

	if err := sess.initialize(ctx); err != nil {
		panic(err)
	}

	if err := sess.virtualGraph.Recover(ctx); err != nil {
		panic(err)
	}

	ticker := time.NewTicker(30 * time.Second)

	sess.state = SessionStateActive

	if sess.parent != nil {
		sess.parent.ReceiveMessage(sessionMessageChildForked{child: sess})
	}

	for {
		select {
		case _, ok := <-sess.stopCh:
			if !ok {
				return
			}

			sess.state = SessionStateClosing

			sess.TryShutdownNow()

		case <-ticker.C:
			if time.Now().Sub(sess.lastKeepAlive) > 30*time.Second {
				sess.RequestShutdown()
			}

		case msg := <-sess.IncomingMessageCh:
			if err := sess.processMessage(ctx, msg); err != nil {
				sess.logger.Error(err)
			}
		}
	}
}

func (sess *Session) processMessage(ctx context.Context, msg coreapi.SessionMessage) error {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	sess.lastKeepAlive = time.Now()

	switch msg := msg.(type) {
	case coreapi.SessionMessageKeepAlive:
		return nil

	case sessionMessageChildForked:
		for _, child := range sess.children {
			if child.UUID() == msg.child.UUID() {
				return nil
			}
		}

		sess.children = append(sess.children, msg.child)

		return nil

	case sessionMessageChildFinished:
		idx := slices.Index(sess.children, msg.child)

		if idx != -1 {
			sess.children = slices.Delete(sess.children, idx, idx+1)
		}

		if sess.state == SessionStateClosing && len(sess.children) == 0 {
			sess.RequestShutdown()
		}

		return nil

	case coreapi.SessionMessageShutdown:
		sess.RequestShutdown()

		return nil

	default:
		logger.Warn("unknown message type: %T", msg)

		return nil
	}
}

func (sess *Session) teardown(ctx context.Context) error {
	if err := sess.serviceProvider.Close(ctx); err != nil {
		return nil
	}

	if err := sess.ClientBusConnection.Close(); err != nil {
		return err
	}

	sess.state = SessionStateClosed

	if sess.parent != nil {
		sess.parent.ReceiveMessage(sessionMessageChildFinished{child: sess})
	}

	return nil
}

func (sess *Session) TryShutdownNow() bool {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	if len(sess.children) > 0 {
		return false
	}

	sess.ShutdownNow()

	return true
}

func (sess *Session) RequestShutdown() {
	sess.stopCh <- struct{}{}
}

func (sess *Session) ShutdownNow() {
	close(sess.stopCh)
}

func (sess *Session) Close() error {
	sess.RequestShutdown()

	return sess.proc.Err()
}

func (sess *Session) BeginTransaction(ctx context.Context, options ...coreapi.TransactionOption) (coreapi.Transaction, error) {
	var err error
	var opts coreapi.TransactionOptions

	opts.Root = sess.config.Root

	opts.Apply(options...)

	tx := &transaction{
		session: sess,
		core:    sess.core,
		opts:    opts,
	}

	tx.sp = inject.NewServiceProvider(
		inject.WithParentServiceProvider(sess.serviceProvider),
	)

	ctx = coreapi.WithSession(ctx, sess)
	ctx = coreapi.WithTransaction(ctx, tx)

	ctx, span := tracer.Start(ctx, "Core.BeginTransaction")
	defer span.End()

	typingManager := inject.Inject[*typing.Manager](tx.sp)
	types := vm.NewTypeRegistry(typingManager)

	tx.lg, err = online.NewLiveGraph(
		ctx,
		opts.Root,
		sess.virtualGraph,
		&sess.lsys,
		types,
		tx.sp,
	)

	if err != nil {
		return nil, err
	}

	inject.Register(tx.sp, func(rctx inject.ResolutionContext) (*vm.Context, error) {
		iso := inject.Inject[*vm.Isolate](rctx)

		return vm.NewContext(ctx, iso, tx.sp), nil
	})

	return tx, nil
}

func (sess *Session) RunTransaction(ctx context.Context, fn coreapi.TransactionFunc, options ...coreapi.TransactionOption) (err error) {
	return coreapi.RunTransaction(ctx, sess, fn, options...)
}

func (sess *Session) initialize(ctx context.Context) error {
	var err error

	sess.mu.Lock()
	defer sess.mu.Unlock()

	if sess.state != SessionStateNew {
		return errors.New("session already initialized")
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

	if sess.config.LinkedStore != nil {
		sess.lsys, err = sess.config.LinkedStore.CreateLinkedStore(ctx, sess.metadataStore)

		if err != nil {
			return err
		}
	} else if sess.parent != nil {
		sess.lsys = sess.parent.lsys
	} else {
		return errors.New("no linked store configured")
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
		&sess.lsys,
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

	sess.state = SessionStateInitialized

	return nil
}

func (sess *Session) initializeServices(ctx context.Context) error {
	inject.Register(sess.serviceProvider, func(ctx inject.ResolutionContext) (*vm.Isolate, error) {
		vms := inject.Inject[*vm.VM](ctx)
		iso := vms.NewIsolate()

		ctx.AppendShutdownHook(func(ctx context.Context) error {
			return iso.Close()
		})

		return iso, nil
	})

	return nil
}
