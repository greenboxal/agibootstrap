package timekeep

import (
	"context"
	"time"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Tick struct {
	X uint64 `json:"x"`
	T uint64 `json:"t"`
}

type TickConsumer interface {
	OnTick(ctx context.Context, tick Tick) error
}

type ITicker interface {
	Start(ctx context.Context, t Tick) error
}

type Ticker struct {
	psi.NodeBase

	Name   string `json:"name"`
	StopAt uint64 `json:"stop_at"`
}

var TickerInterface = psi.DefineNodeInterface[ITicker]()
var TickConsumerInterface = psi.DefineNodeInterface[TickConsumer]()

var TickerType = psi.DefineNodeType[*Ticker](
	psi.WithInterfaceFromNode(TickerInterface),
	psi.WithInterfaceFromNode(TickConsumerInterface),
)

func NewTicker() *Ticker {
	t := &Ticker{}
	t.Init(t)

	return t
}

func (t *Ticker) PsiNodeName() string { return t.Name }

func (t *Ticker) Start(ctx context.Context, tick Tick) error {
	return t.postNextTick(ctx, tick)
}

func (t *Ticker) OnTick(ctx context.Context, tick Tick) error {
	if t.StopAt != 0 && tick.X >= t.StopAt {
		return nil
	}

	return t.postNextTick(ctx, Tick{
		X: tick.X + 1,
		T: uint64(time.Now().UnixNano()),
	})
}

func (t *Ticker) postNextTick(ctx context.Context, tick Tick) error {
	tx := coreapi.GetTransaction(ctx)

	return tx.Notify(ctx, psi.Notification{
		Notifier:  t.CanonicalPath(),
		Notified:  t.CanonicalPath(),
		Interface: TickConsumerInterface.Name(),
		Action:    "OnTick",
		Argument:  tick,
	})
}
