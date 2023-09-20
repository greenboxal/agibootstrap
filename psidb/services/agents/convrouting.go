package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering/themes"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

func (c *Conversation) OnMessageReceived(ctx context.Context, req *OnMessageReceivedRequest) (err error) {
	tx := coreapi.GetTransaction(ctx)
	opts := c.BuildDefaultOptions().MergeWith(req.Options)

	lastMessage, err := req.Message.Resolve(ctx)

	if err != nil {
		return err
	}

	fork := c

	if c.BaseConversation.IsEmpty() || lastMessage.From.Role == msn.RoleUser {
		fork, err = c.Fork(ctx, lastMessage, req.Options)

		if err != nil {
			return err
		}
	}

	ctx = psi.AppendTraceTags(ctx, fork.TraceTags...)

	handleError := func(cause error, dispatch bool) error {
		m := chat.NewMessage(chat.MessageKindError)
		m.From.ID = c.CanonicalPath()
		m.From.Name = "InspectNode"
		m.From.Role = msn.RoleFunction
		m.Text = fmt.Sprintf("Error: %s", cause)

		if req.ToolSelection != nil {
			m.From.Name = req.ToolSelection.Name
		}

		if _, err := fork.addMessage(ctx, m); err != nil {
			return err
		}

		if dispatch {
			if err := fork.dispatchModel(ctx, c.CanonicalPath(), OnMessageReceivedRequest{
				Message:       stdlib.Ref(m),
				Options:       req.Options,
				ToolSelection: req.ToolSelection,
			}); err != nil {
				return err
			}
		}

		return nil
	}

	defer func() {
		if err != nil {
			if err := handleError(err, false); err != nil {
				panic(err)
			}
		}

		if err := c.Update(ctx); err != nil {
			panic(err)
		}
	}()

	pb := NewPromptBuilder()
	pb.WithClient(c.Client)
	pb.WithModelOptions(opts)

	messages, err := fork.SliceMessages(ctx, nil, lastMessage)

	if err != nil {
		return err
	}

	pb.AppendMessage(PromptBuilderHookHistory, messages...)
	pb.SetFocus(lastMessage)

	pb.AppendModelMessage(PromptBuilderHookGlobalSystem, openai.ChatCompletionMessage{
		Role: "system",
		Content: fmt.Sprintf(`You are interfacing with a tree-structured database. The database contains nodes, and each node has a specific NodeType. Depending on its NodeType, a node can have different actions available to it.

QmYXZ is the root of the database. Any path like QmYXZ//foo/bar/baz will be resolved to QmYXZ//foo/bar/baz.
Depending on the NodeType of the selected node, you can perform specific actions on it.
You can read and write files by calling functions on the node. Consult the documentation for more information about the available functions and actions. Follow their declared JSONSchema.
Write messages for the user in Markdown.
Consult the documentation for more information about the available functions and actions. Follow their declared JSONSchema.
The user will send you prompts. These prompts might either be questions related to nodes or direct commands for you to carry out. Your goal is to understand the user's request and utilize the tools and actions at your disposal to satisfy their needs.
`),
	})

	if len(lastMessage.Attachments) > 0 {
		pb.AppendModelMessage(PromptBuilderHookPreFocus, openai.ChatCompletionMessage{
			Role: "system",
			Content: fmt.Sprintf(`
I noticed that your message contains attachments. To manipulate them, please refer to the available functions in the documentation.
`),
		})

		for _, attachment := range lastMessage.Attachments {
			node, err := tx.Resolve(ctx, attachment)

			if err != nil {
				return err
			}

			actions := buildActionsFor(node)

			for _, action := range actions {
				def := openai.FunctionDefinition{
					Name:        strings.Replace(action.Name, ".", "_", -1),
					Description: action.Description,
					Parameters:  action.Parameters,
				}

				pb.WithTools(WrapTool(&def))
			}

			writer := &bytes.Buffer{}

			if err := rendering.RenderNodeWithTheme(ctx, writer, themes.GlobalTheme, "text/markdown", "", node); err != nil {
				return err
			}

			pb.AppendModelMessage(PromptBuilderHookPreFocus, openai.ChatCompletionMessage{
				Role: "system",
				Content: fmt.Sprintf(`You have attached a file. Here are the details:
File Path: %s
File Type: %s

%s`, node.CanonicalPath().String(), node.PsiNodeType().Name(), writer.String()),
			})
		}

		if req.ToolSelection != nil && req.ToolSelection.Name != "" {
			pb.ForceTool(req.ToolSelection.Name)
		} else if req.Options.ForceFunctionCall != nil {
			tool := *req.Options.ForceFunctionCall

			if tool != "" {
				pb.ForceTool(tool)
			}
		} else if lastMessage.From.Role == msn.RoleUser {
			if strings.HasPrefix(lastMessage.Text, "/enter-run-loop") {
				ffc := "ShowAvailableFunctionsForNode"
				pb.ForceTool(ffc)
			}
		}
	}

	args := FunctionCallArgumentHolder{
		Choices: make([]json.RawMessage, 1),
	}

	argsParser := ParseJsonStreamed(PromptParserTargetFunctionCall, args.Prepare(1)...)

	result, err := pb.Execute(ctx, ExecuteWithStreamingParser(argsParser))

	if err != nil {
		return handleError(err, true)
	}

	for _, choice := range result.Choices {
		if err := argsParser.ParseChoice(ctx, choice); err != nil {
			return handleError(err, true)
		}

		arg := args.Choices[0]

		if arg != nil {
			data, err := arg.MarshalJSON()

			if err != nil {
				return handleError(err, true)
			}

			choice.Tool.Arguments = string(data)
		}

		if err := fork.consumeChoice(ctx, lastMessage, choice); err != nil {
			return err
		}
	}

	return c.Update(ctx)
}
