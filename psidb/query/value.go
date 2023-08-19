package query

import "github.com/greenboxal/agibootstrap/pkg/psi"

type Value struct {
	psi.NodeBase

	Value any `json:"value"`
}

type List struct {
	psi.NodeBase

	Values []any `json:"values"`
}
