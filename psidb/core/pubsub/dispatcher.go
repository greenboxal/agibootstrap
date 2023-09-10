package pubsub

import (
	"context"

	"go.uber.org/fx"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type Dispatcher struct {
	pm   *Manager
	core coreapi.Core

	stream *coreapi.ReplicationStreamProcessor
}

func NewDispatcher(
	lc fx.Lifecycle,
	core coreapi.Core,
	pm *Manager,
) *Dispatcher {
	d := &Dispatcher{
		pm:   pm,
		core: core,
	}

	lc.Append(fx.Hook{
		OnStart: d.OnStart,
		OnStop:  d.OnStop,
	})

	return d
}

func (d *Dispatcher) processReplicationMessage(ctx context.Context, entry []*coreapi.JournalEntry) error {
	for _, e := range entry {
		switch e.Op {
		case coreapi.JournalOpWrite:
		case coreapi.JournalOpSetEdge:
		case coreapi.JournalOpRemoveEdge:
		default:
			continue
		}

		if e.Path == nil {
			continue
		}

		if err := d.pm.Publish(ctx, Notification{
			Ts:   e.Ts,
			Path: *e.Path,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (d *Dispatcher) OnStart(ctx context.Context) error {
	slot, err := d.core.CreateReplicationSlot(ctx, coreapi.ReplicationSlotOptions{
		Name:       "core-pubsub",
		Persistent: false,
	})

	if err != nil {
		return err
	}

	d.stream = coreapi.NewReplicationStream(slot, d.processReplicationMessage)

	return nil
}

func (d *Dispatcher) OnStop(ctx context.Context) error {
	return d.stream.Close(ctx)
}
