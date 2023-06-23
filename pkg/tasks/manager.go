package tasks

import (
	"context"
	"sync"

	"github.com/go-errors/errors"
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

var taskCtxKey = struct{ name string }{name: "TaskCtxKey"}

type Manager struct {
	psi.NodeLikeBase

	mu    sync.RWMutex
	tasks map[string]Task
}

func NewManager() *Manager {
	m := &Manager{
		tasks: map[string]Task{},
	}

	m.NodeBase.Init(&m.NodeBase, "")

	return m
}

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

	m.tasks[t.UUID()] = t

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

	delete(m.tasks, t.UUID())

	if t.Parent() == m.PsiNode() {
		t.SetParent(nil)
	}
}
