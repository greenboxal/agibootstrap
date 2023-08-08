package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

type Agent struct {
	psi.NodeBase

	Name string `json:"name"`
}

var AgentType = psi.DefineNodeType[*Agent](
	psi.WithInterfaceFromNode(chat.TopicSubscriberInterface),
)

func (a *Agent) PsiNodeName() string { return a.Name }

func (a *Agent) HandleTopicMessage(ctx context.Context, message *chat.Message) (bool, error) {
	//TODO implement me
	panic("implement me")
}
