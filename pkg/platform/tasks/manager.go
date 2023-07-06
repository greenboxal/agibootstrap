package tasks

import (
	"context"
	"log"
	"sync"

	"github.com/go-errors/errors"
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

var taskCtxKey = struct{ name string }{name: "TaskCtxKey"}

type Manager struct {
	psi.NodeBase

	mu    sync.RWMutex
	tasks map[string]Task
}

var ManagerType = psi.DefineNodeType[*Manager](psi.WithRuntimeOnly())

func NewManager() *Manager {
	m := &Manager{
		tasks: map[string]Task{},
	}

	m.NodeBase.Init(m, psi.WithNodeType(ManagerType))

	return m
}

func (m *Manager) PsiNodeName() string { return "TaskManager" }

func (m *Manager) SpawnTask(ctx context.Context, taskFn TaskFunc) Task {
	parentTaskValue := ctx.Value(taskCtxKey)

	tc := &taskContext{
		done: make(chan struct{}),
	}

	t := &task{
		tc:          tc,
		name:        "",
		description: "",
	}

	tc.t = t

	t.Init()

	tc.ctx, tc.cancel = context.WithCancel(ctx)
	tc.ctx = context.WithValue(tc.ctx, taskCtxKey, t)

	m.mu.Lock()
	defer m.mu.Unlock()

	if parentTask, ok := parentTaskValue.(*task); ok {
		t.SetParent(parentTask)
	} else {
		t.SetParent(m.PsiNode())
	}

	m.tasks[t.uuid] = t

	parent := goprocessctx.WithContext(ctx)
	tc.proc = goprocess.GoChild(parent, func(proc goprocess.Process) {
		var err error

		defer func() {
			_ = proc.CloseAfterChildren()
		}()

		defer tc.Cancel()

		defer tc.onComplete()

		defer func() {
			if e := recover(); err != nil {
				if e, ok := e.(error); ok {
					tc.err = e
				} else {
					tc.err = errors.Wrap(e, 1)
				}

				log.Printf("Task %s failed: %s", t.CanonicalPath(), tc.err.Error())
			}
		}()

		if err = taskFn.Run(tc); err != nil {
			tc.err = err
		}
	})

	go func() {
		<-tc.ctx.Done()

		m.onTaskComplete(t)
	}()

	return t
}

func (m *Manager) onTaskComplete(t *task) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.tasks, t.uuid)

	if t.Parent() == m.PsiNode() {
		t.SetParent(nil)
	}
}
