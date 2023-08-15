package agents

import (
	"context"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	chat2 "github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
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

	if msg.From.Equals(a.CanonicalPath()) {
		return nil
	}

	return a.Step(ctx, msg, message.Path.Parent())
}

func (a *Agent) Step(ctx context.Context, focus psi.Node, replyTo psi.Path) error {
	var options []llm.PredictOption

	tx := coreapi.GetTransaction(ctx)

	topic, err := psi.Resolve[*chat.Topic](ctx, tx.Graph(), replyTo)

	if err != nil {
		return err
	}

	var messages []*chat.Message
	gptMsg := chat2.Message{}

	gptMsg.Entries = append(gptMsg.Entries, chat2.MessageEntry{
		Name: a.Name,
		Role: msn.RoleAI,
		Text: `
Your name is ` + a.Name + `.
`,
	})

	for edges := topic.Edges(); edges.Next(); {
		edge := edges.Value()

		if edge.Kind() == psi.EdgeKindChild {
			to, err := edge.ResolveTo(ctx)

			if err != nil {
				return err
			}

			msg, ok := to.(*chat.Message)

			if !ok {
				continue
			}

			gptMsg.Entries = append(gptMsg.Entries, chat2.MessageEntry{
				Name: msg.From.Name().Name,
				Text: msg.Content,
				Role: msg.Role,
			})

			messages = append(messages, msg)
		}
	}

	result, err := gpt.GlobalModel.PredictChat(ctx, gptMsg, options...)

	if err != nil {
		return err
	}

	replyReq := &chat.PostMessageRequest{
		From:    a.CanonicalPath(),
		Content: result.Entries[0].Text,
		Role:    msn.RoleAI,
	}

	return tx.Notify(ctx, psi.Notification{
		Notifier:  a.CanonicalPath(),
		Notified:  replyTo,
		Interface: chat.TopicInterface.Name(),
		Action:    "PostMessage",
		Argument:  replyReq,
	})
}
