package agents

import (
	"context"

	`github.com/greenboxal/aip/aip-langchain/pkg/providers/openai`

	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

type UpdateTitleRequest struct {
	LastMessage *stdlib.Reference[*chat.Message] `json:"last_message"`
}

func (c *Conversation) UpdateTitle(ctx context.Context, req *UpdateTitleRequest) error {
	if c.Title != "" && !c.IsTitleTemporary {
		return nil
	}

	if c.Title == "" {
		c.IsTitleTemporary = true
	} else {
		c.IsTitleTemporary = false
	}

	lastMessage, err := req.LastMessage.Resolve(ctx)

	if err != nil {
		return err
	}

	pb := NewPromptBuilder()
	pb.WithClient(c.Client)
	pb.WithModelOptions(c.BuildDefaultOptions())

	messages, err := c.SliceMessages(ctx, nil, lastMessage)

	if err != nil {
		return err
	}

	pb.AppendMessage(PromptBuilderHookHistory, messages...)
	pb.SetFocus(lastMessage)

	pb.AppendModelMessage(PromptBuilderHookPostFocus, openai.ChatCompletionMessage{
		Role: "user",
		Content: `
You must reply to this message with the title of the conversation in JSON format (e.g. {"title": "My Conversation"}).
What is the title of this conversation?`,
	})

	var titleStruct struct {
		Title string `json:"title"`
	}

	if err := pb.ExecuteAndParse(ctx, ParseJsonWithPrototype(&titleStruct)); err != nil {
		return nil
	}

	c.Title = titleStruct.Title

	c.Invalidate()
	return c.Update(ctx)
}
