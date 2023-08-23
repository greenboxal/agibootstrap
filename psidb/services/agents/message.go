package agents

import (
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type UserHandle struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Role msn.Role `json:"role"`
}

type Message struct {
	psi.NodeBase

	Timestamp   string         `json:"timestamp"`
	From        UserHandle     `json:"from"`
	Text        string         `json:"text"`
	Attachments []psi.Path     `json:"attachments"`
	Metadata    map[string]any `json:"metadata"`
}

var MessageType = psi.DefineNodeType[*Message]()

func (m *Message) PsiNodeName() string { return m.Timestamp }

func NewMessage() *Message {
	m := &Message{}
	m.Init(m)

	return m
}
