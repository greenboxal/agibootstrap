package chat

import (
	"context"

	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/services/migrations"
)

type Service struct {
	core coreapi.Core
}

func NewService(
	lc fx.Lifecycle,
	core coreapi.Core,
	migrator *migrations.Manager,
) *Service {
	svc := &Service{
		core: core,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return migrator.Migrate(ctx, migrationSet)
		},
	})

	return svc
}

func (s *Service) SendMessage(ctx context.Context, topic psi.Path, message *Message) error {
	return s.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
		topicNode, err := psi.ResolveOrCreate(ctx, tx.Graph(), topic, func() *Topic {
			t := &Topic{Name: topic.Name().Name}
			t.Init(t)
			return t
		})

		if err != nil {
			return err
		}

		message.SetParent(topicNode)

		return topicNode.Update(ctx)
	})
}

func (s *Service) GetMessages(ctx context.Context, topic psi.Path) ([]*Message, error) {
	var messages []*Message

	err := s.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
		t, err := psi.Resolve[*Topic](ctx, tx.Graph(), topic)

		if err != nil {
			return err
		}

		for edges := t.Edges(); edges.Next(); {
			edge := edges.Value()

			if edge.Kind() == psi.EdgeKindChild {
				to, err := edge.ResolveTo(ctx)

				if err != nil {
					return err
				}

				msg, ok := to.(*Message)

				if !ok {
					continue
				}

				messages = append(messages, msg)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return messages, err
}
