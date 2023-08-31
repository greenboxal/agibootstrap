package scheduler

import (
	"context"
	"sync"

	"github.com/alitto/pond"
	"go.uber.org/fx"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type Dispatcher interface {
	DispatchTask(task *Task) error
}

type Scheduler struct {
	ctx    context.Context
	cancel context.CancelFunc

	pool *pond.WorkerPool

	mu     sync.RWMutex
	queues map[string]*TaskQueue
}

func NewScheduler(
	lc fx.Lifecycle,
	config *coreapi.Config,
) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	sch := &Scheduler{
		ctx:    ctx,
		cancel: cancel,

		pool:   config.Workers.Build(),
		queues: map[string]*TaskQueue{},
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return sch.Shutdown(ctx)
		},
	})

	return sch
}

func (s *Scheduler) BaseContext() context.Context {
	return s.ctx
}

func (s *Scheduler) Shutdown(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		s.pool.StopAndWait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		s.pool.Stop()

		return ctx.Err()

	case <-done:
		return nil
	}
}

func (s *Scheduler) ScheduleNext(fn func()) {
	s.pool.Submit(fn)
}

func (s *Scheduler) ScheduleTask(task *Task) {
	qkey := task.entry.Notification.Notified.String()

	s.mu.Lock()
	queue := s.queues[qkey]
	if queue == nil {
		queue = NewTaskQueue()
		queue.Lock()
		s.queues[qkey] = queue
		s.mu.Unlock()
	} else {
		queue.Lock()
		s.mu.Unlock()
	}
	mustSchedule := queue.IsEmpty()
	queue.EnqueueUnlocked(task)
	queue.Unlock()

	if mustSchedule {
		s.scheduleQueue(qkey)
	}
}

func (s *Scheduler) scheduleQueue(qkey string) {
	s.pool.Submit(func() {
		s.mu.RLock()
		queue := s.queues[qkey]
		if queue == nil {
			s.mu.RUnlock()
			return
		}
		task := queue.Dequeue()
		s.mu.RUnlock()

		if task == nil {
			return
		}

		not := task.entry.Notification
		logger.Infow("scheduling task", "handle", task.handle, "queue", not.Notified, "iface", not.Interface, "action", not.Action)

		status, ok := task.TrySchedule()

		if ok && status == coreapi.TaskStatusRunning {
			if err := task.DispatchNow(); err != nil {
				task.OnComplete(err)
			}
		}
	})
}

func (s *Scheduler) notifyWaiter(sc coreapi.Semaphore, waiter coreapi.CompletionWaiter) {
	s.pool.Submit(func() {
		waiter.OnCompleted(sc)
	})
}
