package typing

import (
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/typesystem"
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
	nameComponents := strings.Split(fullName, ".")
	name := nameComponents[len(nameComponents)-1]

	t := &Type{
		Name:     name,
		FullName: fullName,
	}

	t.Init(t, psi.WithNodeType(TypeType))

	return t
}

func (t *Type) PsiNodeName() string { return t.Name }
