package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

type Agent struct {
	psi.NodeBase

	Name        string `json:"name"`
	LastMessage string `json:"last_message"`
}

var AgentType = psi.DefineNodeType[*Agent](
	psi.WithInterfaceFromNode(chat.TopicSubscriberInterface),
)

func (a *Agent) PsiNodeName() string { return a.Name }

func (a *Agent) HandleTopicMessage(ctx context.Context, message *stdlib.Reference[*chat.Message]) error {
	msg, err := message.Resolve(ctx)

	if err != nil {
		return err
	}

	return a.Step(ctx, msg)
}

func (a *Agent) Step(ctx context.Context, focus psi.Node) error {
	a.LastMessage = focus.CanonicalPath().String()

	return a.Update(ctx)
}
