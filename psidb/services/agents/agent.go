package agents

import (
	"context"
	"fmt"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	chat2 "github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/invopop/jsonschema"
	"github.com/samber/lo"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
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

	if msg.From.ID.Equals(a.CanonicalPath()) {
		return nil
	}

	return a.Step(ctx, msg, message.Path.Parent())
}

func (a *Agent) Step(ctx context.Context, focus psi.Node, replyTo psi.Path) error {
	var options []llm.PredictOption

	options = append(options, llm.WithFunctions(buildActionsFor(focus)))

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
				Name: msg.From.Name,
				Text: msg.Text,
				Role: msg.From.Role,
			})

			messages = append(messages, msg)
		}
	}

	result, err := gpt.GlobalModel.PredictChat(ctx, gptMsg, options...)

	if err != nil {
		return err
	}

	replyReq := &chat.SendMessageRequest{
		From: chat.UserHandle{
			ID:   a.CanonicalPath(),
			Role: msn.RoleAI,
		},
		Text: result.Entries[0].Text,
	}

	return tx.Notify(ctx, psi.Notification{
		Notifier:  a.CanonicalPath(),
		Notified:  replyTo,
		Interface: chat.ChatInterface.Name(),
		Action:    "PostMessage",
		Argument:  replyReq,
	})
}

func buildActionsFor(node psi.Node) map[string]llm.FunctionDeclaration {
	var functions []llm.FunctionDeclaration

	for _, iface := range node.PsiNodeType().Interfaces() {
		for _, action := range iface.Interface().Actions() {
			var args llm.FunctionParams

			args.Type = "object"

			if action.RequestType != nil {
				def := action.RequestType.JsonSchema()

				if def.Type == "object" {
					args.Type = def.Type
					args.Required = def.Required
					args.Properties = map[string]*jsonschema.Schema{}

					for _, key := range def.Properties.Keys() {
						v, _ := def.Properties.Get(key)

						args.Properties[key] = v.(*jsonschema.Schema)
					}
				} else {
					args.Type = "object"
					args.Properties = map[string]*jsonschema.Schema{
						"request": def,
					}
					args.Required = []string{"request"}
				}
			} else {
				args.Type = "object"
				args.Properties = map[string]*jsonschema.Schema{
					"request": {
						Type: "string",
					},
				}
				args.Required = []string{"request"}
			}

			functions = append(functions, llm.FunctionDeclaration{
				Name:        fmt.Sprintf("%s_QZQZ_%s", iface.Interface().Name(), action.Name),
				Description: "",
				Parameters:  &args,
			})
		}
	}

	return lo.Associate(functions, func(item llm.FunctionDeclaration) (string, llm.FunctionDeclaration) {
		return item.Name, item
	})
}
