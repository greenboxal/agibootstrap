package promptml

import (
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type ChatMessage struct {
	ContainerBase

	From obsfx.StringProperty
	Role obsfx.SimpleProperty[msn.Role]

	UserData any
}

func Message(from string, role msn.Role, content Node) *ChatMessage {
	cm := &ChatMessage{}

	cm.Init(cm)

	cm.From.SetValue(from)
	cm.Role.SetValue(role)

	cm.AddChildNode(content)

	return cm
}

func MessageWithUserData(from string, role msn.Role, content Node, userData any) *ChatMessage {
	cm := Message(from, role, content)

	cm.UserData = userData

	return cm
}

func MessageWithData(from, to obsfx.ObservableValue[string], role obsfx.ObservableValue[msn.Role], content Node) *ChatMessage {
	cm := &ChatMessage{}

	cm.Init(cm)

	cm.From.Bind(from)
	cm.Role.Bind(role)

	cm.AddChildNode(content)

	return cm
}

type placeholder struct {
	ContainerBase
}

func Placeholder(content ...Node) Parent {
	p := &placeholder{}

	p.Init(p)

	for _, content := range content {
		p.AddChildNode(content)
	}

	return p
}

type childrenIteratorConsumer struct {
	NodeLike

	src iterators.Iterator[AttachableNodeLike]
}

func (c *childrenIteratorConsumer) SetParent(parent psi.Node) {
	for c.src.Next() {
		c.src.Value().SetParent(c.PmlNode())
	}

	c.PmlNode().SetParent(parent)
}

func Map[T any](src iterators.Iterator[T], mapper func(T) AttachableNodeLike) AttachableNodeLike {
	return &childrenIteratorConsumer{
		NodeLike: Container(),

		src: iterators.Map(src, mapper),
	}
}
