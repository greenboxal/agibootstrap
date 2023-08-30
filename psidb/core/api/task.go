package coreapi

import (
	"context"
)

type Semaphore interface {
	IsComplete() bool
	AddWaiter(waiter CompletionWaiter)
	AddWaiterSlot(waiter *CompletionWaiterSlot)
	WaitForCompletion(ctx context.Context) error

	Acquire(value uint64)
	Release(value uint64) bool
}

type CompletionWaiterSlot = ListHead[CompletionWaiter]

type CompletionWaiter interface {
	OnCompleted(sc Semaphore)
}

type CompletionWaiterFunc func(sc Semaphore)

func (s CompletionWaiterFunc) OnCompleted(sc Semaphore) { s(sc) }

type CompletionCondition struct {
	Completion Semaphore
	Value      uint64
}

type TaskHandle struct {
	Xid   uint64 `json:"xid,omitempty"`
	Rid   uint64 `json:"rid,omitempty"`
	Nonce uint64 `json:"nonce,omitempty"`
}

type TaskStatus int

const (
	TaskStatusIdle TaskStatus = iota
	TaskStatusWaiting
	TaskStatusReady
	TaskStatusRunning
	TaskStatusComplete
	TaskStatusCanceled
)
