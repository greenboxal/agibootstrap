package typing

import "github.com/greenboxal/agibootstrap/psidb/psi"

type ActionDefinition struct {
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	RequestType   *psi.Path `json:"request_type,omitempty"`
	ResponseType  *psi.Path `json:"response_type,omitempty"`
	BoundFunction string    `json:"bound_function"`
}

type InterfaceDefinition struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Actions     []ActionDefinition `json:"actions"`
	Module      *psi.Path          `json:"module"`
}
