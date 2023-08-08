package pubsub

import "github.com/greenboxal/agibootstrap/pkg/psi"

type PersistentSubscription struct {
	psi.NodeBase

	SubscriptionID string `json:"id"`

	Topic psi.Path `json:"topic"`
	Depth int      `json:"depth"`

	Subscriber psi.Path `json:"subscriber"`
}

var PersistentSubscriptionType = psi.DefineNodeType[*PersistentSubscription]()

func (s *PersistentSubscription) PsiNodeName() string { return s.SubscriptionID }
