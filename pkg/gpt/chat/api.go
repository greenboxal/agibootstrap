package chat

import (
	"context"
)

// CreateChannelRequest defines the information needed to create a new chat channel.
type CreateChannelRequest struct {
	Name string `json:"name"`
}

// CreateChannelResponse defines the response from creating a new chat channel.
type CreateChannelResponse struct {
	Channel Channel `json:"channel"`
}

// GetChannelHistoryRequest defines the request to get the message history of a channel.
type GetChannelHistoryRequest struct {
	ChannelID string `json:"channel_id"`
}

// GetChannelHistoryResponse defines the response from getting the message history of a channel.
type GetChannelHistoryResponse struct {
	Messages []Message `json:"messages"`
}

// PostMessageRequest defines the request to post a message in a channel.
type PostMessageRequest struct {
	ChannelID string `json:"channel_id"`
	Content   string `json:"content"`
}

// PostMessageResponse defines the response from posting a message in a channel.
type PostMessageResponse struct {
	Message Message `json:"message"`
}

// SubscribeRequest defines the request to subscribe to a channel.
type SubscribeRequest struct {
	ChannelID string `json:"channel_id"`
}

// SubscribeResponse defines the response from subscribing to a channel.
type SubscribeResponse struct {
	SubscriptionID string `json:"subscription_id"`
}

// Service defines the interface for the chat service.
type Service interface {
	CreateChannel(ctx context.Context, req CreateChannelRequest) (*CreateChannelResponse, error)
	GetChannelHistory(ctx context.Context, req GetChannelHistoryRequest) (*GetChannelHistoryResponse, error)
	PostMessage(ctx context.Context, req PostMessageRequest) (*PostMessageResponse, error)
	Subscribe(ctx context.Context, req SubscribeRequest) (*SubscribeResponse, error)
}
