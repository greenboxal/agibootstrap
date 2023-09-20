package agents

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"golang.org/x/exp/slices"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

func (c *Conversation) GetMessages(ctx context.Context, req *chat.GetMessagesRequest) ([]*chat.Message, error) {
	var err error
	var from, to *chat.Message
	var messages []*chat.Message

	if !req.From.IsEmpty() {
		from, err = req.From.Resolve(ctx)

		if err != nil {
			return nil, err
		}
	}

	if !req.To.IsEmpty() {
		to, err = req.To.Resolve(ctx)

		if err != nil {
			return nil, err
		}
	}

	if !req.SkipBaseHistory && c.BaseConversation != nil && (from == nil || from.Parent() != c) {
		baseConversation, err := c.BaseConversation.Resolve(ctx)

		if err != nil {
			return nil, err
		}

		baseMessage, err := c.BaseMessage.Resolve(ctx)

		if err != nil {
			return nil, err
		}

		baseMessages, err := baseConversation.SliceMessages(ctx, from, baseMessage)

		if err != nil {
			return nil, err
		}

		messages = append(messages, baseMessages...)
	}

	ownMessages := psi.GetEdges(c, chat.ConversationMessageEdge)

	ownMessages = lo.Filter(ownMessages, func(item *chat.Message, index int) bool {
		if from != nil && item.Timestamp < from.Timestamp {
			return false
		}

		if to != nil && item.Timestamp > to.Timestamp {
			return false
		}

		return true
	})

	slices.SortFunc(ownMessages, func(i, j *chat.Message) bool {
		return strings.Compare(i.Timestamp, j.Timestamp) < 0
	})

	messages = append(messages, ownMessages...)

	return messages, nil
}

func (c *Conversation) SliceMessages(ctx context.Context, from, to *chat.Message) ([]*chat.Message, error) {
	msgs, err := c.GetMessages(ctx, &chat.GetMessagesRequest{
		From: stdlib.Ref(from),
		To:   stdlib.Ref(to),
	})

	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (c *Conversation) SendMessage(ctx context.Context, req *chat.SendMessageRequest) (*chat.Message, error) {
	tx := coreapi.GetTransaction(ctx)

	m := chat.NewMessage(chat.MessageKindEmit)
	m.Timestamp = ""
	m.From = req.From
	m.Text = req.Text
	m.Attachments = req.Attachments
	m.Metadata = req.Metadata

	if req.Function != "" {
		m.FunctionCall = &chat.FunctionCall{
			Name:      req.Function,
			Arguments: req.FunctionArguments,
		}
	}

	m, err := c.addMessage(ctx, m)

	if err != nil {
		return nil, err
	}

	if err := tx.Notify(ctx, psi.Notification{
		Notifier:  c.CanonicalPath(),
		Notified:  c.CanonicalPath(),
		Interface: ConversationInterface.Name(),
		Action:    "OnMessageReceived",
		Argument: &OnMessageReceivedRequest{
			Message: stdlib.Ref(m),
			Options: req.ModelOptions,
		},
	}); err != nil {
		return nil, err
	}

	return m, c.Update(ctx)
}

func (c *Conversation) addMessage(ctx context.Context, m *chat.Message) (*chat.Message, error) {
	if m.Timestamp == "" {
		m.Timestamp = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	if err := m.Update(ctx); err != nil {
		return nil, err
	}

	if m.Parent() == nil {
		if m.Kind == chat.MessageKindError {
			logger.Error(m.Text)
		}

		m.SetParent(c)
	}

	c.SetEdge(chat.ConversationMessageEdge.Named(m.Timestamp), m)

	return m, c.Update(ctx)
}
