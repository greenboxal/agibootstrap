package turing

import "github.com/greenboxal/agibootstrap/psidb/typesystem"

type ValueKind int

const (
	ValueKindInvalid ValueKind = iota
	ValueKindError
	ValueKindObject
	ValueKindString
	ValueKindNumber
	ValueKindReference
)

type Value struct {
	Kind  ValueKind
	Value typesystem.Value
}
