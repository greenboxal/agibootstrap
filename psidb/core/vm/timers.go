package vm

import (
	"context"
	"time"
)

type Timers interface {
	SetTimeout(fn func(), ms int) int
	ClearTimeout(id int)

	SetInterval(fn func(), ms int) int
	ClearInterval(id int)
}

type basicTimers struct {
	ctx     *Context
	timers  map[int]func()
	counter int
}

func (b *basicTimers) SetTimeout(fn func(), ms int) int {
	b.counter++

	id := b.counter

	ctx, cancel := context.WithCancel(b.ctx.baseCtx)
	timer := time.After(time.Duration(ms) * time.Millisecond)

	go func() {
		defer delete(b.timers, id)

		select {
		case <-ctx.Done():
			return
		case <-timer:
			fn()
		}
	}()

	b.timers[id] = cancel

	return id
}

func (b *basicTimers) ClearTimeout(id int) {
	cancel := b.timers[id]

	if cancel != nil {
		cancel()
	}
}

func (b *basicTimers) SetInterval(fn func(), ms int) int {
	b.counter++

	id := b.counter

	ctx, cancel := context.WithCancel(b.ctx.baseCtx)
	ticker := time.NewTicker(time.Duration(ms) * time.Millisecond)

	go func() {
		defer delete(b.timers, id)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn()
		}
	}()

	b.timers[id] = cancel

	return id
}

func (b *basicTimers) ClearInterval(id int) {
	cancel := b.timers[id]

	if cancel != nil {
		cancel()
	}
}
