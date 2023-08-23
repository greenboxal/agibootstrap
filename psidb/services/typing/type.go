package typing

import (
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
	Fields        []FieldDefinition        `json:"fields,omitempty"`
	Interfaces    []InterfaceDefinition    `json:"interfaces,omitempty"`
	PrimitiveKind typesystem.PrimitiveKind `json:"primitive_kind,omitempty"`

	Schema jsonschema.Schema `json:"-"`
}

var TypeType = psi.DefineNodeType[*Type]()

func (t *Type) PsiNodeName() string { return t.Name }
