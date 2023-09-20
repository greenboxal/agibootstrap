package triggers

import (
	"context"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type TriggerHandler interface {
	HandleTriggerActivation(ctx context.Context, act Activation) error
}

type Activation struct {
	Target *stdlib.Reference[psi.Node] `json:"target"`
}

type Trigger struct {
	psi.NodeBase

	Name     string   `json:"name"`
	Notified psi.Path `json:"notified"`
	IsSync   bool     `json:"is_sync"`
}

func (t *Trigger) PsiNodeName() string { return t.Name }

func (t *Trigger) Dispatch(ctx context.Context, act Activation) error {
	not := psi.Notification{
		Notifier:  t.CanonicalPath(),
		Notified:  t.Notified,
		Interface: TriggerHandlerInterface.Name(),
		Action:    "HandleTriggerActivation",
		Argument:  act,
	}

	if t.IsSync {
		target, err := coreapi.GetTransaction(ctx).Resolve(ctx, t.Notified)

		if err != nil {
			return err
		}

		_, err = not.Apply(ctx, target)

		return err
	}

	return coreapi.Dispatch(ctx, not)
}

func AddTrigger(node psi.Node, name string) *Trigger {
	trigger := &Trigger{
		Name: name,
	}

	trigger.Init(trigger)

	node.SetEdge(TriggerEdge.Named(name), trigger)

	return trigger
}

func GetTriggers(node psi.Node) []*Trigger {
	return psi.GetEdges(node, TriggerEdge)
}

func GetTrigger(node psi.Node, name string) *Trigger {
	return psi.GetEdgeOrNil[*Trigger](node, TriggerEdge.Named(name))
}

func DispatchTriggers(ctx context.Context, target psi.Node, act Activation) error {
	for n := target; n != nil; n = n.Parent() {
		for _, trigger := range GetTriggers(n) {
			if err := trigger.Dispatch(ctx, act); err != nil {
				return err
			}
		}
	}

	return nil
}

var TriggerHandlerInterface = psi.DefineNodeInterface[TriggerHandler]()
var TriggerType = psi.DefineNodeType[*Trigger]()
var TriggerEdge = psi.DefineEdgeType[*Trigger]("psidb.trigger")
