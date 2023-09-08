package chat

import (
	"context"

	"go.uber.org/fx"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/migrations"
)

type Service struct {
	core coreapi.Core
}

func NewService(
	lc fx.Lifecycle,
	core coreapi.Core,
	migrator migrations.Migrator,
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

func (s *Service) SendMessage(ctx context.Context, path psi.Path, req *SendMessageRequest) error {
	return s.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
		topic, err := psi.ResolveOrCreate[*Topic](ctx, tx.Graph(), path, func() *Topic {
			t := &Topic{Name: path.Name().Name}
			t.Init(t)
			return t
		})

		if err != nil {
			return err
		}

		_, err = topic.SendMessage(ctx, req)

		if err != nil {
			return err
		}

		return nil
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
