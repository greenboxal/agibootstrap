package kb

import (
	"fmt"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Page struct {
	psi.NodeBase

	Title string `json:"title"`
	Body  string `json:"body"`
}

var PageType = psi.DefineNodeType[*Page]()

func NewPage() *Page {
	p := &Page{}
	p.Init(p)

	return p
}

func (p *Page) PsiNodeName() string { return fmt.Sprintf("Page-%d", p.Title) }
