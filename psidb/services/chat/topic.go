package chat

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
)

type TopicSubscriber interface {
	HandleTopicMessage(ctx context.Context, message *stdlib.Reference[*Message]) error
}

var TopicSubscriberInterface = psi.DefineNodeInterface[TopicSubscriber]()

type Topic struct {
	psi.NodeBase

	Name    string     `json:"name"`
	Members []psi.Path `json:"members"`
}

var TopicType = psi.DefineNodeType[*Topic](
	psi.WithInterfaceFromNode(ChatInterface),
	psi.WithInterfaceFromNode(TopicSubscriberInterface),
)

func (t *Topic) PsiNodeName() string { return t.Name }

func (t *Topic) GetMessages(ctx context.Context, req *GetMessagesRequest) ([]*Message, error) {
	var err error
	var from, to *Message
	var messages []*Message

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

	ownMessages := psi.GetEdges(t, ConversationMessageEdge)

	ownMessages = lo.Filter(ownMessages, func(item *Message, index int) bool {
		if from != nil && item.Timestamp < from.Timestamp {
			return false
		}

		if to != nil && item.Timestamp > to.Timestamp {
			return false
		}

		return true
	})

	slices.SortFunc(ownMessages, func(i, j *Message) bool {
		return strings.Compare(i.Timestamp, j.Timestamp) < 0
	})

	messages = append(messages, ownMessages...)

	return messages, nil
}

func (t *Topic) SendMessage(ctx context.Context, req *SendMessageRequest) (*Message, error) {
	m := NewMessage(MessageKindEmit)
	m.Timestamp = ""
	m.From = req.From
	m.Text = req.Text
	m.Attachments = req.Attachments
	m.Metadata = req.Metadata

	if req.Function != "" {
		m.FunctionCall = &FunctionCall{
			Name:      req.Function,
			Arguments: req.FunctionArguments,
		}
	}

	if m.Timestamp == "" {
		m.Timestamp = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	m.ReplyTo = append(m.ReplyTo, t.CanonicalPath())

	if m.Parent() == nil {
		m.SetParent(t)
	}

	if err := m.Update(ctx); err != nil {

		return nil, err
	}

	t.SetEdge(ConversationMessageEdge.Named(m.Timestamp), m)

	if err := coreapi.Dispatch(ctx, psi.Notification{
		Notifier:  t.CanonicalPath(),
		Notified:  t.CanonicalPath(),
		Interface: TopicSubscriberInterface.Name(),
		Action:    "HandleTopicMessage",
		Argument:  stdlib.Ref(m),
	}); err != nil {
		return nil, err
	}

	return m, t.Update(ctx)
}

func (t *Topic) HandleTopicMessage(ctx context.Context, ref *stdlib.Reference[*Message]) error {
	tx := coreapi.GetTransaction(ctx)

	m, err := ref.Resolve(ctx)

	if err != nil {
		return err
	}

	t.SetEdge(ConversationMessageEdge.Named(m.Timestamp), m)

	for _, member := range t.Members {
		if member.Equals(t.CanonicalPath()) {
			continue
		}

		if member.Equals(m.From.ID) {
			continue
		}

		if err := tx.Notify(ctx, psi.Notification{
			Notifier:  t.CanonicalPath(),
			Notified:  member,
			Interface: TopicSubscriberInterface.Name(),
			Action:    "HandleTopicMessage",
			Argument:  ref,
		}); err != nil {
			return err
		}
	}

	return t.Update(ctx)
}
