package chat

import "github.com/greenboxal/agibootstrap/pkg/psi"

type Message struct {
	psi.NodeBase

	Ts      string `json:"ts"`
	Content string `json:"content"`
}

var MessageType = psi.DefineNodeType[*Message]()

func (m *Message) PsiNodeName() string { return m.Ts }
