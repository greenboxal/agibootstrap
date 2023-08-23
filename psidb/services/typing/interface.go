package typing

import "github.com/greenboxal/agibootstrap/pkg/psi"

type ActionDefinition struct {
	Name          string    `json:"name"`
	RequestType   *psi.Path `json:"request_type,omitempty"`
	ResponseType  *psi.Path `json:"response_type,omitempty"`
	BoundFunction string    `json:"bound_function"`
}

type InterfaceDefinition struct {
	Name    string             `json:"name"`
	Actions []ActionDefinition `json:"actions"`
	Module  *psi.Path          `json:"module"`
}
