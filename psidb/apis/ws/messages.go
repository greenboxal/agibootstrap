package ws

import (
	"github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/core/pubsub"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type Message struct {
	Mid     uint64 `json:"mid"`
	ReplyTo uint64 `json:"reply_to"`

	Ack         *AckMessage          `json:"ack,omitempty"`
	Nack        *NackMessage         `json:"nack,omitempty"`
	Error       *ErrorMessage        `json:"error,omitempty"`
	Subscribe   *SubscribeMessage    `json:"subscribe,omitempty"`
	Unsubscribe *UnsubscribeMessage  `json:"unsubscribe,omitempty"`
	Notify      *NotificationMessage `json:"notify,omitempty"`
	Session     *SessionMessage      `json:"session,omitempty"`
}

var MessageType = typesystem.TypeOf(Message{})

type AckMessage struct {
}

type NackMessage struct {
}

type SessionMessage struct {
	SessionID string                 `json:"session_id"`
	Message   coreapi.SessionMessage `json:"message"`
}

type NotificationMessage struct {
	Notification pubsub.Notification `json:"notification"`
}

type SubscribeMessage struct {
	Topic string `json:"topic"`
	Depth int    `json:"depth"`
}

type UnsubscribeMessage struct {
	Topic string `json:"topic"`
}

type ErrorMessage struct {
	Error string `json:"error"`
}
