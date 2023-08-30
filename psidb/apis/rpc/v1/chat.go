package rpcv1

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

type Chat struct {
	svc *chat.Service
}

func NewChat(svc *chat.Service) *Chat {
	return &Chat{svc: svc}
}

type SendMessageRequest struct {
	Topic          *stdlib.Reference[*chat.Topic] `json:"topic"`
	MessageRequest chat.SendMessageRequest        `json:"request"`
}

type SendMessageResponse struct {
}

func (c *Chat) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	if err := c.svc.SendMessage(ctx, req.Topic.Path, &req.MessageRequest); err != nil {
		return nil, err
	}

	return &SendMessageResponse{}, nil
}

type GetMessagesRequest struct {
	Topic psi.Path `json:"topic"`
}

type GetMessagesResponse struct {
	Messages []*chat.Message `json:"messages"`
}

func (c *Chat) GetMessages(ctx context.Context, req *GetMessagesRequest) (*GetMessagesResponse, error) {
	msgs, err := c.svc.GetMessages(ctx, req.Topic)

	if err != nil {
		return nil, err
	}

	return &GetMessagesResponse{Messages: msgs}, nil
}
