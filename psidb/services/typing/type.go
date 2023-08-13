package typing

import (
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type ActionDefinition struct {
	Name          string `json:"name"`
	RequestType   string `json:"request_type"`
	ResponseType  string `json:"response_type"`
	BoundFunction string `json:"bound_function"`
}

type InterfaceDefinition struct {
	Name    string             `json:"name"`
	Actions []ActionDefinition `json:"actions"`
	Module  *psi.Path          `json:"module"`
}

type FieldDefinition struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Type struct {
	psi.NodeBase

	Name       string                `json:"name"`
	Fields     []FieldDefinition     `json:"fields,omitempty"`
	Interfaces []InterfaceDefinition `json:"interfaces,omitempty"`
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
