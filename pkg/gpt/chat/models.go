package chat

import (
	"time"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

// Message represents a single message in a channel.
type Message struct {
	psi.NodeBase

	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Channel represents a chat channel.
type Channel struct {
	psi.NodeBase

	ID   string `json:"id"`
	Name string `json:"name"`
}

var MessageType = psi.DefineNodeType[*Message]()
var ChannelType = psi.DefineNodeType[*Channel]()
