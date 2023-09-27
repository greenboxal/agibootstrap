package typing

import (
	"github.com/invopop/jsonschema"

	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type FieldDefinition struct {
	Name string   `json:"name"`
	Type psi.Path `json:"type"`
}

type Type struct {
	psi.NodeBase

	Name          string                   `json:"name"`
	FullName      string                   `json:"full_name"`
	Fields        []FieldDefinition        `json:"fields,omitempty"`
	Interfaces    []InterfaceDefinition    `json:"interfaces,omitempty"`
	PrimitiveKind typesystem.PrimitiveKind `json:"primitive_kind,omitempty"`

	Schema *jsonschema.Schema `json:"-"`
}

var TypeType = psi.DefineNodeType[*Type]()

func NewType(fullName string) *Type {
	nameComponents := typesystem.ParseTypeName(fullName)

	t := &Type{
		Name:     nameComponents.NameWithArgs(),
		FullName: fullName,
	}

	t.Init(t, psi.WithNodeType(TypeType))

	return t
}

func (t *Type) PsiNodeName() string { return t.Name }
