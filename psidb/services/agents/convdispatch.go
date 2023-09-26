package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/invopop/jsonschema"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	openai2 "github.com/sashabaranov/go-openai"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering/themes"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

func (c *Conversation) AcceptChoice(ctx context.Context, baseMessage *chat.Message, choice PromptResponseChoice) error {
	return c.consumeChoice(ctx, baseMessage, choice)
}

func (c *Conversation) consumeChoice(ctx context.Context, baseMessage *chat.Message, choice PromptResponseChoice) error {
	m, err := c.addMessage(ctx, choice.Message)

	if err != nil {
		return err
	}

	if choice.Reason == openai2.FinishReasonFunctionCall || choice.Tool != nil && choice.Tool.Name != "" {
		if (choice.Tool.Focus == nil || choice.Tool.Focus.IsEmpty()) && len(baseMessage.Attachments) > 0 {
			if choice.Tool == nil {
				choice.Tool = &PromptToolSelection{}
			}

			choice.Tool.Focus = stdlib.RefFromPath[psi.Node](baseMessage.Attachments[0])
		}

		if err := c.dispatchSideEffect(ctx, c.CanonicalPath(), OnMessageSideEffectRequest{
			Message:       stdlib.Ref(m),
			Options:       c.BuildDefaultOptions(),
			ToolSelection: choice.Tool,
		}); err != nil {
			return err
		}
	} else if c.BaseConversation != nil {
		if err := c.Merge(ctx, m); err != nil {
			return err
		}
	}

	return coreapi.DispatchSelf(ctx, c, ConversationInterface, "UpdateTitle", &UpdateTitleRequest{
		LastMessage: stdlib.Ref(m),
	})
}

func (c *Conversation) OnMessageSideEffect(ctx context.Context, req *OnMessageSideEffectRequest) (err error) {
	ctx = psi.AppendTraceTags(ctx, c.TraceTags...)
	tx := coreapi.GetTransaction(ctx)

	handleError := func(cause error, dispatch bool) error {
		m := chat.NewMessage(chat.MessageKindError)
		m.From.ID = c.CanonicalPath()
		m.From.Name = "InspectNode"
		m.From.Role = msn.RoleFunction
		m.Text = fmt.Sprintf("Error: %s", cause)

		if req.ToolSelection != nil {
			m.From.Name = req.ToolSelection.Name
		}

		if _, err := c.addMessage(ctx, m); err != nil {
			return err
		}

		if dispatch {
			if err := c.dispatchModel(ctx, c.CanonicalPath(), OnMessageReceivedRequest{
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
			if err := handleError(err, true); err != nil {
				panic(err)
			}
		}
	}()

	baseMessage, err := req.Message.Resolve(ctx)

	if err != nil {
		return err
	}

	switch req.ToolSelection.Name {
	case "CallNodeAction":
		var args struct {
			Path      psi.Path        `json:"path"`
			ToolName  string          `json:"tool_name"`
			Arguments json.RawMessage `json:"arguments"`
		}

		if err := json.Unmarshal([]byte(req.ToolSelection.Arguments), &args); err != nil {
			panic(err)
		}

		rawArgs, err := args.Arguments.MarshalJSON()

		if err != nil {
			panic(err)
		}

		req.ToolSelection.Arguments = string(rawArgs)
		req.ToolSelection.Name = args.ToolName

		if !args.Path.IsEmpty() {
			req.ToolSelection.Focus = stdlib.RefFromPath[psi.Node](args.Path)
		}

	case "ShowAvailableFunctionsForNode":
		fallthrough
	case "ActionForNode":
		var args struct {
			Path psi.Path `json:"path"`
		}

		if strings.TrimSpace(req.ToolSelection.Arguments) != "" {
			if err := json.Unmarshal([]byte(req.ToolSelection.Arguments), &args); err != nil {
				panic(err)
			}
		}

		if args.Path.IsEmpty() && len(baseMessage.Attachments) > 0 {
			args.Path = baseMessage.Attachments[0]
		}

		target, err := tx.Resolve(ctx, args.Path)

		if err != nil {
			return handleError(err, true)
		}

		return c.showAvailableActions(ctx, req, target)

	case "TraverseToNode":
		fallthrough
	case "TraverseTo":
		fallthrough
	case "InspectNode":
		var args struct {
			Path psi.Path `json:"path"`
		}

		if err := json.Unmarshal([]byte(req.ToolSelection.Arguments), &args); err != nil {
			panic(err)
		}

		if args.Path.IsEmpty() && len(baseMessage.Attachments) > 0 {
			args.Path = baseMessage.Attachments[0]
		}

		target, err := tx.Resolve(ctx, args.Path)

		if err != nil {
			return handleError(err, true)
		}

		return c.inspectNode(ctx, req, target)
	}

	return c.dispatchNodeAction(ctx, req)
}

func (c *Conversation) dispatchNodeAction(ctx context.Context, req *OnMessageSideEffectRequest) error {
	target, err := req.ToolSelection.Focus.Resolve(ctx)

	if err != nil {
		return err
	}

	ifaceName, actionName, _ := strings.Cut(req.ToolSelection.Name, "_QZQZ_")

	iface := target.PsiNodeType().Interface(ifaceName)

	if iface == nil {
		return fmt.Errorf("interface %s not found", ifaceName)
	}

	action := iface.Action(actionName)

	if action == nil {
		return fmt.Errorf("action %s not found", actionName)
	}

	/*if action.RequestType() != nil {
		schema := action.RequestType().JsonSchema()
		bundle := typesystem.Universe().CompileBundleFor(schema)
		form := NewForm(bundle)

		if err := form.ParseJson([]byte(req.ToolSelection.Arguments)); err != nil {
			return handleError(err, true)
		}

		valid, err := form.Validate()

		if err != nil {
			return handleError(err, true)
		}

		if !valid || req.ToolSelection.Arguments == "" {
			tctx := c.CreateThreadContext(ctx, baseMessage)
			tctx, err = tctx.Fork(ctx, baseMessage, req.Options)

			if err != nil {
				return err
			}

			if req.ToolSelection.Arguments == "" {
				if err := form.FillAll(tctx); err != nil {
					return handleError(err, true)
				}
			} else {
				if ok, err := form.Fix(tctx); err != nil {
					return handleError(err, true)
				} else if !ok {
					return c.dispatchSideEffect(ctx, c.CanonicalPath(), *req)
				}
			}
		}

		fixed, err := form.ToJSON()

		if err != nil {
			return err
		}

		req.ToolSelection.Arguments = string(fixed)
	}*/

	not := psi.Notification{
		Notifier:  c.CanonicalPath(),
		Notified:  target.CanonicalPath(),
		Interface: ifaceName,
		Action:    actionName,
		Params:    []byte(req.ToolSelection.Arguments),
	}

	writer := &bytes.Buffer{}
	attachments := []psi.Path{target.CanonicalPath()}

	func() {
		defer func() {
			if err := recover(); err != nil {
				writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
			}
		}()

		result, err := not.Apply(ctx, target)

		if err != nil {
			writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
			return
		}

		if node, ok := result.(psi.Node); ok {
			if node.Parent() == nil {
				node.SetParent(target)

				if err := node.Update(ctx); err != nil {
					writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
					return
				}
			}

			attachments = append(attachments, node.CanonicalPath())

			if err := rendering.RenderNodeWithTheme(ctx, writer, themes.GlobalTheme, "text/markdown", "", node); err != nil {
				writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
				return
			}
		} else if it, ok := result.(psi.EdgeIterator); ok {
			for it.Next() {
				edge := it.Value()
				node := edge.To()

				attachments = append(attachments, node.CanonicalPath())

				if err := rendering.RenderNodeWithTheme(ctx, writer, themes.GlobalTheme, "text/markdown", "", node); err != nil {
					writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
					return
				}
			}
		} else if it, ok := result.(psi.NodeIterator); ok {
			for it.Next() {
				node := it.Value()

				attachments = append(attachments, node.CanonicalPath())

				if err := rendering.RenderNodeWithTheme(ctx, writer, themes.GlobalTheme, "text/markdown", "", node); err != nil {
					writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
					return
				}
			}
		} else if result != nil {
			if err := ipld.EncodeStreaming(writer, typesystem.Wrap(result), dagjson.Encode); err != nil {
				writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
				return
			}
		}

		writer.Write([]byte(`Please proceed to the next step of the operation you requested. What specific action would you like me to take now?`))
	}()

	replyMessage := chat.NewMessage(chat.MessageKindEmit)
	replyMessage.From.ID = c.CanonicalPath()
	replyMessage.From.Name = req.ToolSelection.Name
	replyMessage.From.Role = msn.RoleFunction
	replyMessage.Text = writer.String()
	replyMessage.Attachments = attachments

	replyMessage, err = c.addMessage(ctx, replyMessage)

	if err != nil {
		return err
	}

	return c.dispatchModel(ctx, c.CanonicalPath(), OnMessageReceivedRequest{
		Message: stdlib.Ref(replyMessage),
		Options: req.Options,
	})
}

func (c *Conversation) showAvailableActions(ctx context.Context, req *OnMessageSideEffectRequest, target psi.Node) error {
	writer := &bytes.Buffer{}
	actionMap := map[string]*jsonschema.Schema{}

	for _, iface := range target.PsiNodeType().Interfaces() {
		for _, action := range iface.Interface().Actions() {
			var schema *jsonschema.Schema

			if action.RequestType != nil {
				schema = action.RequestType.JsonSchema()
			}

			if schema == nil {
				schema = &jsonschema.Schema{Type: "object"}
			}

			actionName := fmt.Sprintf("%s_QZQZ_%s", iface.Interface().Name(), action.Name)
			actionMap[actionName] = schema
		}
	}

	_, _ = fmt.Fprintf(writer, "**Path:** %s\n", target.CanonicalPath().String())
	_, _ = fmt.Fprintf(writer, "**Node Type:** %s\n", target.PsiNodeType().Name())
	_, _ = fmt.Fprintf(writer, "# Actions\n\n```json\n")

	if err := json.NewEncoder(writer).Encode(actionMap); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(writer, "\n```\n")

	replyMessage := chat.NewMessage(chat.MessageKindEmit)
	replyMessage.From.ID = c.CanonicalPath()
	replyMessage.From.Name = req.ToolSelection.Name
	replyMessage.From.Role = msn.RoleFunction
	replyMessage.Text = writer.String()
	replyMessage.Attachments = []psi.Path{target.CanonicalPath()}

	replyMessage, err := c.addMessage(ctx, replyMessage)

	if err != nil {
		return err
	}

	return c.dispatchModel(ctx, c.CanonicalPath(), OnMessageReceivedRequest{
		Message: stdlib.Ref(replyMessage),
		Options: req.Options,
	})
}

func (c *Conversation) inspectNode(ctx context.Context, req *OnMessageSideEffectRequest, target psi.Node) error {
	tx := coreapi.GetTransaction(ctx)
	writer := &bytes.Buffer{}

	_, _ = fmt.Fprintf(writer, "**Path:** %s\n", target.CanonicalPath().String())
	_, _ = fmt.Fprintf(writer, "**Node Type:** %s\n", target.PsiNodeType().Name())

	_, _ = fmt.Fprintf(writer, "# Edges\n\n")
	for edges := target.Edges(); edges.Next(); {
		edge := edges.Value()
		to, err := edge.ResolveTo(ctx)

		if err != nil {
			writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
		} else {
			_, _ = fmt.Fprintf(writer, "- **%s:** %s\n", edge.Key(), to.PsiNodeType().Name())
		}
	}

	_, _ = fmt.Fprintf(writer, "# Node\n\n")
	if err := rendering.RenderNodeWithTheme(ctx, writer, themes.GlobalTheme, "text/markdown", "", target); err != nil {
		writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
	}

	replyMessage := chat.NewMessage(chat.MessageKindEmit)
	replyMessage.From.ID = c.CanonicalPath()
	replyMessage.From.Name = req.ToolSelection.Name
	replyMessage.From.Role = msn.RoleFunction
	replyMessage.Text = writer.String()
	replyMessage.Attachments = []psi.Path{target.CanonicalPath()}

	replyMessage, err := c.addMessage(ctx, replyMessage)

	if err != nil {
		return err
	}

	return tx.Notify(ctx, psi.Notification{
		Notifier:  c.CanonicalPath(),
		Notified:  c.CanonicalPath(),
		Interface: ConversationInterface.Name(),
		Action:    "OnMessageReceived",
		Argument: &OnMessageReceivedRequest{
			Message: stdlib.Ref(replyMessage),
			Options: req.Options,
		},
	})
}

func (c *Conversation) dispatchModel(ctx context.Context, requestor psi.Path, request OnMessageReceivedRequest) error {
	tx := coreapi.GetTransaction(ctx)

	if err := tx.Notify(ctx, psi.Notification{
		Notifier:  requestor,
		Notified:  c.CanonicalPath(),
		Interface: ConversationInterface.Name(),
		Action:    "OnMessageReceived",
		Argument:  request,
	}); err != nil {
		return err
	}

	return nil
}

func (c *Conversation) dispatchSideEffect(ctx context.Context, path psi.Path, request OnMessageSideEffectRequest) error {
	tx := coreapi.GetTransaction(ctx)

	if err := tx.Notify(ctx, psi.Notification{
		Notifier:  path,
		Notified:  c.CanonicalPath(),
		Interface: ConversationInterface.Name(),
		Action:    "OnMessageSideEffect",
		Argument:  request,
	}); err != nil {
		return err
	}

	return nil
}
