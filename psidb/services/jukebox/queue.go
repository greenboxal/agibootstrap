package jukebox

import (
	"context"
	"sync"
	"time"

	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type CommandWithTimeCode struct {
	Index    int
	TimeCode int

	EvaluableCommand
}

type PlayableItemState struct {
	Item PlayerQueueItem

	Commands []CommandWithTimeCode
	Position int

	Done bool

	Connection *Connection
}

func (ps *PlayableItemState) Reset(item PlayerQueueItem) {
	ps.Item = item
	ps.Position = 0
	ps.Commands = item.BuildCommands()
	ps.Done = false
	ps.Connection = nil
}

func (ps *PlayableItemState) SeekTimeCode(time int) {
	for i, cmd := range ps.Commands {
		if cmd.TimeCode >= time {
			ps.Position = i

			return
		}
	}
}

func (ps *PlayableItemState) SeekIndex(index int) {
	if index < 0 {
		index = 0
	} else if index >= len(ps.Commands) {
		index = len(ps.Commands)
	}

	ps.Position = index
}

func (ps *PlayableItemState) NextCommand() (CommandWithTimeCode, bool) {
	if ps.Done {
		return CommandWithTimeCode{}, false
	}

	index := ps.Position

	if index >= len(ps.Commands) {
		ps.Done = true
		return CommandWithTimeCode{}, false
	}

	ps.Position++

	return ps.Commands[index], true
}

type PlayerQueueItem struct {
	ID       psi.Path
	Context  *CommandContext
	Commands []EvaluableCommand
}

func (i PlayerQueueItem) BuildCommands() []CommandWithTimeCode {
	result := make([]CommandWithTimeCode, len(i.Commands))

	for i, cmd := range i.Commands {
		ctc := CommandWithTimeCode{
			Index:            i,
			TimeCode:         i,
			EvaluableCommand: cmd,
		}

		result[i] = ctc
	}

	return result
}

type PlayerQueue struct {
	manager *QueueManager
	name    psi.Path

	mu   sync.RWMutex
	cond *sync.Cond

	items []PlayerQueueItem
	state PlayableItemState

	proc goprocess.Process

	isPlaying bool
}

func NewPlayerQueue(manager *QueueManager, name psi.Path) *PlayerQueue {
	q := &PlayerQueue{
		manager: manager,
		name:    name,

		items: make([]PlayerQueueItem, 0, 256),
	}

	q.cond = sync.NewCond(&q.mu)
	q.proc = goprocess.Go(q.run)

	return q
}

func (q *PlayerQueue) Play(ctx context.Context) {
	q.mu.Lock()

	if q.isPlaying {
		q.mu.Unlock()
		return
	}

	q.isPlaying = true
	q.cond.Broadcast()
	q.mu.Unlock()

	q.updateState(ctx)
}

func (q *PlayerQueue) Pause(ctx context.Context) {
	q.mu.Lock()

	if !q.isPlaying {
		q.mu.Unlock()
		return
	}

	q.isPlaying = false
	q.mu.Unlock()

	q.updateState(ctx)
}

func (q *PlayerQueue) Seek(ctx context.Context, time int) {
	defer q.updateState(ctx)

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.state.Item.ID.IsEmpty() {
		return
	}

	q.state.SeekTimeCode(time)
}

func (q *PlayerQueue) Enqueue(item PlayerQueueItem) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = append(q.items, item)
	q.cond.Broadcast()
}

func (q *PlayerQueue) Remove(id psi.Path) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, item := range q.items {
		if item.ID.Equals(id) {
			q.items = append(q.items[:i], q.items[i+1:]...)
		}
	}
}

func (q *PlayerQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = nil
}

func (q *PlayerQueue) popNextItem(wait bool) *PlayerQueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.items) == 0 {
		if !wait {
			return nil
		}

		q.cond.Wait()
	}

	item := q.items[0]
	q.items = q.items[1:]

	return &item
}

func (q *PlayerQueue) run(proc goprocess.Process) {
	proc.Go(func(proc goprocess.Process) {
		ctx := goprocessctx.OnClosingContext(proc)
		ticker := time.NewTicker(750 * time.Millisecond)

		for {
			select {
			case <-proc.Closing():
				return
			case <-ticker.C:
				if q.isPlaying {
					q.updateState(ctx)
				}
			}
		}
	})

	for {
		select {
		case <-proc.Closing():
			return
		default:
		}

		if !q.isPlaying {
			q.mu.Lock()
			for !q.isPlaying {
				q.cond.Wait()
			}
			q.mu.Unlock()
		}

		select {
		case <-proc.Closing():
			return
		default:
		}

		if q.state.Item.ID.IsEmpty() || q.state.Done {
			item := q.popNextItem(true)

			q.state.Reset(*item)
		}

		select {
		case <-proc.Closing():
			return
		default:
		}

		if q.state.Connection == nil {
			q.state.Connection = q.manager.cm.GetOrCreateConnection(q.state.Item.Context.Channel)
			q.state.Item.Context.Out = q.state.Connection.out
		}

		cmd, ok := q.state.NextCommand()

		if !ok {
			continue
		}

		if err := cmd.Evaluate(q.state.Item.Context); err != nil {
			logger.Error(err)
		}

		select {
		case <-proc.Closing():
			return
		default:
		}
	}
}

func (q *PlayerQueue) Close() error {
	return q.proc.Close()
}

func (q *PlayerQueue) updateState(ctx context.Context) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	err := q.manager.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
		n, err := tx.Resolve(ctx, q.name)

		if err != nil {
			return err
		}

		state := n.(*Player)

		state.IsPlaying = q.isPlaying
		state.CurrentItem = q.state.Item.ID
		state.CurrentTimeCode = q.state.Position

		state.Invalidate()
		return state.Update(ctx)
	})

	if err != nil {
		panic(err)
	}
}
