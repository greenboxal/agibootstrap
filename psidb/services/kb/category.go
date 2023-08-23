package kb

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type ICategory interface {
	Consolidate(ctx context.Context) error
}

type Category struct {
	psi.NodeBase

	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var HasCategoryEdge = psi.DefineEdgeType[*Category]("kb.category")
var HasDocument = psi.DefineEdgeType[*Document]("kb.document")
var CategoryInterface = psi.DefineNodeInterface[ICategory]()
var CategoryType = psi.DefineNodeType[*Category](psi.WithInterfaceFromNode(CategoryInterface))

func NewCategory(slug string) *Category {
	p := &Category{}
	p.Slug = slug
	p.Init(p, psi.WithNodeType(CategoryType))

	return p
}

func (cat *Category) PsiNodeName() string { return cat.Slug }

func (cat *Category) AddDocument(doc *Document) {
	cat.SetEdge(HasDocument.Named(doc.Slug), doc)
	doc.SetEdge(HasCategoryEdge.Named(cat.Slug), cat)
}

func (cat *Category) Consolidate(ctx context.Context) error {

	return nil
}
