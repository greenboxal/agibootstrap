package chat

import (
	"context"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
)

type TopicSubscriber interface {
	HandleTopicMessage(ctx context.Context, self psi.Node, message *stdlib.Reference[*Message]) error
}

type ITopic interface {
	PostMessage(ctx context.Context, self psi.Node, req *PostMessageRequest) (*Message, error)
}

var TopicSubscriberInterface = psi.DefineNodeInterface[TopicSubscriber]()
var TopicInterface = psi.DefineNodeInterface[ITopic]()

type Topic struct {
	psi.NodeBase

	Name    string     `json:"name"`
	Members []psi.Path `json:"members"`
}

var TopicType = psi.DefineNodeType[*Topic](
	psi.WithInterfaceFromNode(TopicInterface),
)

func (t *Topic) PsiNodeName() string { return t.Name }

func (t *Topic) PostMessage(ctx context.Context, req *PostMessageRequest) (*Message, error) {
	msg := NewMessage()
	msg.Role = req.Role
	msg.From = req.From
	msg.Content = req.Content

	t.AddChildNode(msg)

	if err := t.Update(ctx); err != nil {
		return nil, err
	}

	return msg, t.HandleTopicMessage(ctx, stdlib.Ref(msg))
}

func (t *Topic) HandleTopicMessage(ctx context.Context, ref *stdlib.Reference[*Message]) error {
	tx := coreapi.GetTransaction(ctx)

	messageData, err := ipld.Encode(typesystem.Wrap(ref), dagjson.Encode)

	if err != nil {
		return nil
	}

	for _, member := range t.Members {
		if err := tx.Notify(ctx, psi.Notification{
			Notifier:  t.CanonicalPath(),
			Notified:  member,
			Interface: TopicSubscriberInterface.Name(),
			Action:    "HandleTopicMessage",
			Params:    messageData,
		}); err != nil {
			return err
		}
	}

	return nil
}
