package rpcv1

import (
	"context"
	"fmt"
	"time"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

type Chat struct {
	svc *chat.Service
}

func NewChat(svc *chat.Service) *Chat {
	return &Chat{svc: svc}
}

type SendMessageRequest struct {
	Topic psi.Path `json:"topic"`

	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type SendMessageResponse struct {
}

func (c *Chat) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	msg := &chat.Message{
		Ts:      fmt.Sprintf("%d", time.Now().UnixNano()),
		Content: req.Message.Content,
	}

	msg.Init(msg)

	if err := c.svc.SendMessage(ctx, req.Topic, msg); err != nil {
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
