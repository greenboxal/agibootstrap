package stdlib

import "github.com/greenboxal/agibootstrap/pkg/psi"

type Text struct {
	psi.NodeBase

	Value string `json:"value"`
}

var TextType = psi.DefineNodeType[*Text]()

func NewText(value string) *Text {
	t := &Text{
		Value: value,
	}

	t.Init(t, psi.WithNodeType(TextType))

	return t
}
