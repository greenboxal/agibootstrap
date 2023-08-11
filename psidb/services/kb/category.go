package kb

import (
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Category struct {
	psi.NodeBase

	Name        string `json:"name"`
	Description string `json:"description"`
}

var CategoryType = psi.DefineNodeType[*Category]()

func NewCategory() *Category {
	p := &Category{}
	p.Init(p)

	return p
}

func (cat *Category) PsiNodeName() string { return cat.Name }
