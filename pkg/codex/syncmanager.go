package codex

import (
	"context"
	"sync"

	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type SyncManagerCommand func(ctx context.Context) error

type SyncManager struct {
	psi.NodeBase

	logger *zap.SugaredLogger
	p      *Project

	mu sync.Mutex

	initialSyncDone chan struct{}
	isSyncing       bool

	proc         goprocess.Process
	commandQueue chan SyncManagerCommand
}

var SyncManagerType = psi.DefineNodeType[*SyncManager](psi.WithRuntimeOnly())

func NewSyncManager(p *Project) *SyncManager {
	sm := &SyncManager{
		logger: logging.GetLogger("codex/syncmanager"),

		p: p,

		commandQueue: make(chan SyncManagerCommand, 100),

		initialSyncDone: make(chan struct{}),
	}

	sm.Init(sm, psi.WithNodeType(SyncManagerType))
	sm.p.g.AddListener(sm)

	sm.proc = goprocess.Go(sm.run)

	return sm
}

func (sm *SyncManager) OnNodeUpdated(node psi.Node) {
	if sm.isSyncing {
		ctx := goprocessctx.OnClosingContext(sm.proc)
		err := sm.syncSubtree(ctx, node)

		if err != nil {
			sm.logger.Error(err)
		}
	} else {
		sm.commandQueue <- func(ctx context.Context) error {
			return sm.syncSubtree(ctx, node)
		}
	}
}

func (sm *SyncManager) syncSubtree(ctx context.Context, node psi.Node) error {

	return nil
}

func (sm *SyncManager) RequestSync(ctx context.Context, runNow bool) error {
	if sm.isSyncing {
		return nil
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.isSyncing {
		return nil
	}

	sm.isSyncing = true

	dispatch := func(ctx context.Context) error {
		defer func() {
			sm.isSyncing = false

			if sm.initialSyncDone != nil {
				close(sm.initialSyncDone)
				sm.initialSyncDone = nil
			}
		}()

		return sm.performSync(ctx)
	}

	if runNow {
		return dispatch(ctx)
	} else {
		sm.commandQueue <- dispatch
	}

	return nil
}

func (sm *SyncManager) performSync(ctx context.Context) error {
	maxDepth := 0
	count := 0

	sm.logger.Infow("Performing sync walk", "root", sm.p.rootPath)

	err := psi.Walk(sm.p.rootNode, func(cursor psi.Cursor, entering bool) error {
		if cursor.Depth() > maxDepth {
			maxDepth = cursor.Depth()
		}

		n := cursor.Value()

		if n, ok := n.(*vfs.Directory); ok && entering {
			count++

			cursor.WalkChildren()

			return n.Sync(func(path string) bool {
				return !sm.p.repo.IsIgnored(path)
			})
		} else {
			cursor.SkipChildren()
		}

		return nil
	})

	if err != nil {
		return err
	}

	sm.logger.Info("Performing sync update")

	return sm.p.Update(ctx)
}

func (sm *SyncManager) run(proc goprocess.Process) {
	ctx := goprocessctx.OnClosingContext(proc)

	for {
		select {
		case <-ctx.Done():
			if err := proc.CloseAfterChildren(); err != nil {
				panic(err)
			}

			return

		case cmd := <-sm.commandQueue:
			func() {
				defer func() {
					if r := recover(); r != nil {
						sm.logger.Error(r)
					}
				}()

				if err := cmd(ctx); err != nil {
					sm.logger.Error(err)
				}
			}()
		}
	}
}

func (sm *SyncManager) Close() error {
	if sm.proc != nil {
		if err := sm.proc.Close(); err != nil {
			return err
		}

		sm.proc = nil
	}

	return nil
}

func (sm *SyncManager) Enqueue(task SyncManagerCommand) {
	sm.commandQueue <- task
}

func (sm *SyncManager) WaitForInitialSync(ctx context.Context) error {
	if sm.initialSyncDone == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case _, _ = <-sm.initialSyncDone:
		return nil
	}
}
