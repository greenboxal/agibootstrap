package agents

import (
	"context"
	"strconv"
	"time"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

func (c *Conversation) ForkAsChatLog(ctx context.Context, baseMessage *chat.Message, options gpt.ModelOptions) (ChatLog, error) {
	fork, err := c.Fork(ctx, baseMessage, options)

	if err != nil {
		return nil, err
	}

	return fork, nil
}

func (c *Conversation) Fork(ctx context.Context, baseMessage *chat.Message, options gpt.ModelOptions) (*Conversation, error) {
	fork := &Conversation{}
	fork.Name = strconv.FormatInt(time.Now().UnixNano(), 10)
	fork.BaseConversation = stdlib.Ref(c)
	fork.BaseMessage = stdlib.Ref(baseMessage)
	fork.BaseOptions = c.BuildDefaultOptions().MergeWith(options)
	fork.TraceTags = append(fork.TraceTags, c.TraceTags...)
	fork.TraceTags = append(fork.TraceTags, c.CanonicalPath().String())
	fork.Init(fork)
	fork.SetParent(c)

	if err := fork.Update(ctx); err != nil {
		return nil, err
	}

	joinMsg := chat.NewMessage(chat.MessageKindJoin)
	joinMsg.Attachments = []psi.Path{fork.CanonicalPath()}

	if _, err := c.addMessage(ctx, joinMsg); err != nil {
		return nil, err
	}

	return fork, nil
}

func (c *Conversation) Merge(ctx context.Context, focus *chat.Message) error {
	var mergeMsg *chat.Message

	baseMsg, err := c.BaseMessage.Resolve(ctx)

	if err != nil {
		return err
	}

	if !c.BaseConversation.IsEmpty() {
		base, err := c.BaseConversation.Resolve(ctx)

		if err != nil {
			return err
		}

		mergeMsg = chat.NewMessage(chat.MessageKindMerge)
		mergeMsg.Attachments = []psi.Path{focus.CanonicalPath(), base.CanonicalPath()}

		if _, err := c.addMessage(ctx, mergeMsg); err != nil {
			return err
		}

		c.IsMerged = true
		c.Invalidate()

		if err := coreapi.Dispatch(ctx, psi.Notification{
			Notifier:  c.CanonicalPath(),
			Notified:  base.CanonicalPath(),
			Interface: ConversationInterface.Name(),
			Action:    "OnForkMerging",
			Argument: OnForkMergingRequest{
				Fork:       stdlib.Ref(c),
				MergePoint: stdlib.Ref(mergeMsg),
			},
		}); err != nil {
			return err
		}
	}

	if len(baseMsg.ReplyTo) > 0 {
		msgs, err := c.SliceMessages(ctx, baseMsg, mergeMsg)

		if err != nil {
			return err
		}

		for _, msg := range msgs {
			if msg.Kind == chat.MessageKindEmit && msg.From.Role == msn.RoleAI && msg.FunctionCall == nil {
				for _, replyTo := range baseMsg.ReplyTo {
					if err := coreapi.Dispatch(ctx, psi.Notification{
						Notifier:  c.CanonicalPath(),
						Notified:  replyTo,
						Interface: chat.TopicSubscriberInterface.Name(),
						Action:    "HandleTopicMessage",
						Argument:  stdlib.Ref(msg),
					}); err != nil {
						return err
					}
				}
			}
		}
	}

	return c.Update(ctx)
}

type OnForkMergingRequest struct {
	Fork       *stdlib.Reference[*Conversation] `json:"fork" jsonschema:"title=Fork,description=The fork to merge"`
	MergePoint *stdlib.Reference[*chat.Message] `json:"merge_point" jsonschema:"title=Merge Point,description=The merge point"`
}

func (c *Conversation) OnForkMerging(ctx context.Context, req *OnForkMergingRequest) (err error) {
	var lastMessage *chat.Message

	ctx = psi.AppendTraceTags(ctx, c.TraceTags...)

	fork, err := req.Fork.Resolve(ctx)

	if err != nil {
		return err
	}

	baseMessage, err := fork.BaseMessage.Resolve(ctx)

	if err != nil {
		return err
	}

	mergePoint, err := req.MergePoint.Resolve(ctx)

	if err != nil {
		return err
	}

	msgs, err := fork.SliceMessages(ctx, baseMessage, mergePoint)

	if err != nil {
		return err
	}

	for _, msg := range msgs {
		if msg.Kind == chat.MessageKindError || (msg.Kind == chat.MessageKindEmit && msg.From.Role == msn.RoleAI && msg.FunctionCall == nil) {
			if _, err := c.addMessage(ctx, msg); err != nil {
				return err
			}
		}

		lastMessage = msg
	}

	if lastMessage == nil {
		lastMessage = mergePoint
	}

	return coreapi.DispatchSelf(ctx, c, ConversationInterface, "UpdateTitle", &UpdateTitleRequest{
		LastMessage: stdlib.Ref(lastMessage),
	})
}
