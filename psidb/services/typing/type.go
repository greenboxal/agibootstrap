package typing

import (
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Type struct {
	psi.NodeBase

	Name string `json:"name"`
}

var TypeType = psi.DefineNodeType[*Type]()

func NewType(name string) *Type {
	t := &Type{
		Name: name,
	}

	t.Init(t, psi.WithNodeType(TypeType))

	return t
}

func (t *Type) PsiNodeName() string { return t.Name }
