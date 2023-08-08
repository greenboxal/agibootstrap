package chat

import (
	"strconv"
	"time"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type PostMessageRequest struct {
	From    psi.Path `json:"from"`
	Content string   `json:"content"`
}

type Message struct {
	psi.NodeBase

	Ts      string   `json:"ts"`
	From    psi.Path `json:"from"`
	Content string   `json:"content"`
}

var MessageType = psi.DefineNodeType[*Message]()

func NewMessage() *Message {
	m := &Message{
		Ts: strconv.FormatUint(uint64(time.Now().UnixNano()), 10),
	}

	m.Init(m)

	return m
}

func (m *Message) PsiNodeName() string { return m.Ts }
