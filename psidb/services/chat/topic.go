package chat

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type TopicSubscriber interface {
	HandleTopicMessage(ctx context.Context, node psi.Node, message *Message) (bool, error)
}

var TopicSubscriberInterface = psi.DefineNodeInterface[TopicSubscriber]()

type Topic struct {
	psi.NodeBase

	Name string `json:"name"`
}

var TopicType = psi.DefineNodeType[*Topic]()

func (m *Topic) PsiNodeName() string { return m.Name }
