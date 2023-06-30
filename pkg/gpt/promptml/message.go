package promptml

import (
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type ChatMessage struct {
	ContainerBase

	From obsfx.StringProperty
	To   obsfx.StringProperty
	Role obsfx.SimpleProperty[msn.Role]
}

func Message(from, to string, role msn.Role, content Node) *ChatMessage {
	cm := &ChatMessage{}

	cm.Init(cm, "")

	cm.From.SetValue(from)
	cm.To.SetValue(to)
	cm.Role.SetValue(role)

	cm.AddChildNode(content)

	return cm
}

func MessageWithData(from, to obsfx.ObservableValue[string], role obsfx.ObservableValue[msn.Role], content Node) *ChatMessage {
	cm := &ChatMessage{}

	cm.Init(cm, "")

	cm.From.Bind(from)
	cm.To.Bind(to)
	cm.Role.Bind(role)

	cm.AddChildNode(content)

	return cm
}
