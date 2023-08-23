package agents

import (
	"context"
	"fmt"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	chat2 "github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/invopop/jsonschema"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"

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

func convertType(typ *jsonschema.Schema) *openai.JSONSchemaDefine {
	result := &openai.JSONSchemaDefine{}

	result.Type = openai.JSONSchemaType(typ.Type)
	result.Description = typ.Description
	result.Required = typ.Required

	if len(typ.Enum) > 0 {
		result.Enum = lo.Map(typ.Enum, func(item interface{}, _ int) string {
			return item.(string)
		})
	}

	if typ.Items != nil {
		result.Items = convertType(typ.Items)
	}

	if typ.Properties != nil {
		result.Properties = map[string]*openai.JSONSchemaDefine{}

		for _, k := range typ.Properties.Keys() {
			v, ok := typ.Properties.Get(k)

			if !ok {
				continue
			}

			result.Properties[k] = convertType(v.(*jsonschema.Schema))
		}
	}

	return result
}

func buildActionsFor(node psi.Node) map[string]llm.FunctionDeclaration {
	var functions []llm.FunctionDeclaration

	for _, iface := range node.PsiNodeType().Interfaces() {
		for _, action := range iface.Interface().Actions() {
			var args llm.FunctionParams

			args.Type = "object"

			if action.RequestType != nil {
				def := convertType(action.RequestType.JsonSchema())
				args.Type = def.Type
				args.Properties = def.Properties
				args.Required = def.Required
			} else {
				args.Type = "object"
				args.Properties = map[string]*openai.JSONSchemaDefine{
					"request": {
						Type: "string",
					},
				}
				args.Required = []string{"request"}
			}

			functions = append(functions, llm.FunctionDeclaration{
				Name:        fmt.Sprintf("%s_%s", iface.Interface().Name(), action.Name),
				Description: "",
				Parameters:  &args,
			})
		}
	}

	return lo.Associate(functions, func(item llm.FunctionDeclaration) (string, llm.FunctionDeclaration) {
		return item.Name, item
	})
}
