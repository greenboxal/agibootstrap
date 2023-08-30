package scheduler

import coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"

type TaskQueue struct {
	coreapi.ListHead[*Task]
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{}
}

func (q *TaskQueue) Enqueue(t *Task) {
	q.Lock()
	defer q.Unlock()

	q.EnqueueUnlocked(t)
}

func (q *TaskQueue) EnqueueUnlocked(t *Task) {
	t.queue.Lock()
	defer t.queue.Unlock()

	t.queue.Value = t
	t.queue.Next = nil

	if q.Prev == nil {
		t.queue.Prev = nil
		q.Prev = &t.queue
	} else {
		q.Prev.Next = &t.queue
		t.queue.Prev = q.Prev
		q.Prev = &t.queue
	}

	if q.Next == nil {
		q.Next = q.Prev
	}
}

func (q *TaskQueue) Dequeue() *Task {
	q.Lock()
	defer q.Unlock()

	return q.DequeueUnlocked()
}
func (q *TaskQueue) DequeueUnlocked() *Task {
	head := q.Next

	if head == nil {
		return nil
	}

	head.Lock()
	defer head.Unlock()

	if q.Prev == head {
		q.Prev = head.Prev
	}

	q.Next = head.Next
	head.Next = nil

	return head.Value
}

func (q *TaskQueue) IsEmpty() bool {
	return q.Next == nil
}
