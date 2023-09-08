package stdlib

import "github.com/greenboxal/agibootstrap/psidb/psi"

type Collection struct {
	psi.NodeBase

	Name string `json:"name"`
}

var CollectionType = psi.DefineNodeType[*Collection]()

func NewCollection(name string) *Collection {
	c := &Collection{
		Name: name,
	}

	c.Init(c)

	return c
}

func (c *Collection) PsiNodeName() string { return c.Name }
