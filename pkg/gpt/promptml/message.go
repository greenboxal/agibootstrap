package promptml

import (
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type ChatMessage struct {
	ContainerBase

	From obsfx.StringProperty
	Role obsfx.SimpleProperty[msn.Role]
}

func Message(from string, role msn.Role, content Node) *ChatMessage {
	cm := &ChatMessage{}

	cm.Init(cm, "")

	cm.From.SetValue(from)
	cm.Role.SetValue(role)

	cm.AddChildNode(content)

	return cm
}

func MessageWithData(from, to obsfx.ObservableValue[string], role obsfx.ObservableValue[msn.Role], content Node) *ChatMessage {
	cm := &ChatMessage{}

	cm.Init(cm, "")

	cm.From.Bind(from)
	cm.Role.Bind(role)

	cm.AddChildNode(content)

	return cm
}
