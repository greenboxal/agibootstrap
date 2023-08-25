package agents

import (
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type UserHandle struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Role msn.Role `json:"role"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type MessageKind string

const (
	MessageKindEmit  MessageKind = "emit"
	MessageKindError MessageKind = "error"
	MessageKindJoin  MessageKind = "join"
	MessageKindMerge MessageKind = "merge"
)

type Message struct {
	psi.NodeBase

	Kind        MessageKind    `json:"kind"`
	Timestamp   string         `json:"timestamp"`
	From        UserHandle     `json:"from"`
	Text        string         `json:"text"`
	Attachments []psi.Path     `json:"attachments"`
	Metadata    map[string]any `json:"metadata"`

	FunctionCall *FunctionCall `json:"function_call"`
}

var MessageType = psi.DefineNodeType[*Message]()

func (m *Message) PsiNodeName() string { return m.Timestamp }

func NewMessage(kind MessageKind) *Message {
	m := &Message{}
	m.Kind = kind
	m.Init(m)

	return m
}

func (m *Message) FromOpenAI(msg openai.ChatCompletionMessage) {
	m.From = UserHandle{
		Name: msg.Name,
		Role: openai.ConvertToRole(msg.Role),
	}

	m.Text = msg.Content

	if msg.FunctionCall != nil {
		m.FunctionCall = &FunctionCall{
			Name:      msg.FunctionCall.Name,
			Arguments: msg.FunctionCall.Arguments,
		}
	}
}

func (m *Message) ToOpenAI() openai.ChatCompletionMessage {
	result := openai.ChatCompletionMessage{}

	result.Name = m.From.Name
	result.Role = openai.ConvertFromRole(m.From.Role)
	result.Content = m.Text

	return result
}
