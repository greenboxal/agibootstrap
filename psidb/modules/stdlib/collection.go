package stdlib

import "github.com/greenboxal/agibootstrap/pkg/psi"

type Collection struct {
	psi.NodeBase

	Name string `json:"name"`
}

var CollectionType = psi.DefineNodeType[*Collection]()

func (c *Collection) PsiNodeName() string {
	return c.Name
}
