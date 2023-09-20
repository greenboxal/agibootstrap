package jukebox

import (
	"context"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type PlayRequest struct{}
type PauseRequest struct{}
type StopRequest struct{}
type SeekRequest struct {
	Time int `json:"time"`
}
type QueueRequest struct {
	Path psi.Path `json:"path"`
}
type DequeueRequest struct {
	Path psi.Path `json:"path"`
}

type IPlayer interface {
	Play(ctx context.Context) error
	Pause(ctx context.Context) error
	Stop(ctx context.Context) error
	Seek(ctx context.Context, req SeekRequest) error
	Queue(ctx context.Context, req QueueRequest) error
	Dequeue(ctx context.Context, req DequeueRequest) error
	DequeueAll(ctx context.Context) error
}

var PlayerInterface = psi.DefineNodeInterface[IPlayer]()

type Player struct {
	psi.NodeBase

	Name string `json:"name"`

	ItemQueue []psi.Path `json:"queue"`

	CurrentItem     psi.Path `json:"current_item"`
	CurrentTimeCode int      `json:"current_time_code"`

	IsPlaying bool `json:"is_playing"`

	QueueManager  *QueueManager `json:"-" inject:""`
	QueueInstance *PlayerQueue  `json:"-"`
}

var _ IPlayer = (*Player)(nil)

var PlayerType = psi.DefineNodeType[*Player](psi.WithInterfaceFromNode(PlayerInterface))

func (p *Player) PsiNodeName() string {
	return p.Name
}

func (p *Player) Play(ctx context.Context) error {
	p.IsPlaying = true
	p.Invalidate()

	p.getQueue().Play(ctx)

	return p.Update(ctx)
}

func (p *Player) Pause(ctx context.Context) error {
	p.IsPlaying = false
	p.Invalidate()

	if err := p.Update(ctx); err != nil {
		return err
	}

	p.getQueue().Pause(ctx)

	return nil
}

func (p *Player) Stop(ctx context.Context) error {
	p.IsPlaying = false
	p.CurrentTimeCode = 0
	p.CurrentItem = psi.Path{}
	p.Invalidate()

	p.getQueue().Pause(ctx)

	return p.Update(ctx)
}

func (p *Player) Seek(ctx context.Context, req SeekRequest) error {
	p.CurrentTimeCode = req.Time
	p.Invalidate()

	p.getQueue().Seek(ctx, req.Time)

	return p.Update(ctx)
}

func (p *Player) Queue(ctx context.Context, req QueueRequest) error {
	p.ItemQueue = append(p.ItemQueue, req.Path)
	p.Invalidate()

	p.enqueue(ctx, req.Path)

	return p.Update(ctx)
}

func (p *Player) Dequeue(ctx context.Context, req DequeueRequest) error {
	for i, path := range p.ItemQueue {
		if path.Equals(req.Path) {
			p.ItemQueue = append(p.ItemQueue[:i], p.ItemQueue[i+1:]...)

			break
		}
	}

	p.Invalidate()

	p.getQueue().Remove(req.Path)

	return p.Update(ctx)
}

func (p *Player) DequeueAll(ctx context.Context) error {
	p.ItemQueue = nil
	p.Invalidate()

	p.getQueue().Clear()

	return p.Update(ctx)
}

func (p *Player) advanceToNextItem(ctx context.Context) error {
	if len(p.ItemQueue) == 0 {
		return nil
	}

	p.CurrentTimeCode = 0
	p.CurrentItem = p.ItemQueue[0]
	p.ItemQueue = p.ItemQueue[1:]
	p.Invalidate()

	p.enqueue(ctx, p.CurrentItem)

	return p.Update(ctx)
}

func (p *Player) OnUpdate(ctx context.Context) error {
	return nil
}

func (p *Player) getQueue() *PlayerQueue {
	if p.QueueInstance == nil {
		p.QueueInstance = p.QueueManager.GetOrCreateQueue(p.CanonicalPath())
	}

	return p.QueueInstance
}

func (p *Player) enqueue(ctx context.Context, path psi.Path) {
	tx := coreapi.GetTransaction(ctx)
	n, err := tx.Resolve(ctx, path)

	if err != nil {
		panic(err)
	}

	item := PlayerQueueItem{
		ID:      path,
		Context: &CommandContext{},
	}

	switch n := n.(type) {
	case *Song:
		item.Commands = n.Commands
	}

	p.getQueue().Enqueue(item)
}
