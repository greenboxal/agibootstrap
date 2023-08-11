package kb

import (
	"fmt"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Entity struct {
	psi.NodeBase

	Index   int    `json:"index"`
	Content string `json:"title"`
}

var EntityType = psi.DefineNodeType[*Entity]()

func NewEntity() *Entity {
	p := &Entity{}
	p.Init(p)

	return p
}

func (p *Entity) PsiNodeName() string { return fmt.Sprintf("Entity-%d", p.Index) }
