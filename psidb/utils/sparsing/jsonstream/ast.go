package jsonstream

import (
	"encoding/json"
	"strconv"

	"github.com/greenboxal/agibootstrap/psidb/typesystem"
	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
)

type Node = sparsing.Node
type IToken = sparsing.IToken

type Value interface {
	Node

	GetPrimitiveType() typesystem.PrimitiveKind
	GetValue() any
}

type Pair struct {
	CompositeNodeBase

	Key   *String
	Colon IToken
	Value Value
}

type Object struct {
	CompositeNodeBase

	Pairs []*Pair
}

func (o *Object) GetValue() any {
	result := make(map[string]any)

	for _, p := range o.Pairs {
		result[p.Key.Value] = p.Value.GetValue()
	}

	return result
}

func (o *Object) GetPrimitiveType() typesystem.PrimitiveKind {
	return typesystem.PrimitiveKindMap
}

type Array struct {
	CompositeNodeBase

	Values []Value
}

func (a *Array) GetPrimitiveType() typesystem.PrimitiveKind {
	return typesystem.PrimitiveKindList
}

func (a *Array) GetValue() any {
	result := make([]any, len(a.Values))

	for i, v := range a.Values {
		result[i] = v.GetValue()
	}

	return result
}

type String struct {
	TerminalNodeBase

	Value string
}

func (s *String) GetPrimitiveType() typesystem.PrimitiveKind {
	return typesystem.PrimitiveKindString
}

func (s *String) GetValue() any { return s.Value }

func (s *String) SetTerminalToken(tk IToken) error {
	s.Start = tk
	s.End = tk
	s.Value = tk.GetValue()

	return json.Unmarshal([]byte(tk.GetValue()), &s.Value)
}

type Number struct {
	TerminalNodeBase

	Value float64
}

func (n *Number) GetPrimitiveType() typesystem.PrimitiveKind {
	return typesystem.PrimitiveKindFloat
}

func (n *Number) GetValue() any { return n.Value }

func (n *Number) SetTerminalToken(tk IToken) error {
	n.Start = tk
	n.End = tk

	val, err := strconv.ParseFloat(tk.GetValue(), 64)

	if err != nil {
		return err
	}

	n.Value = val

	return nil
}

type Boolean struct {
	TerminalNodeBase

	Value bool
}

func (b *Boolean) GetPrimitiveType() typesystem.PrimitiveKind {
	return typesystem.PrimitiveKindBoolean
}

func (b *Boolean) GetValue() any { return b.Value }

func (b *Boolean) SetTerminalToken(tk IToken) error {
	b.Start = tk
	b.End = tk

	switch tk.GetValue() {
	case "true":
		b.Value = true
	case "false":
		b.Value = false
	default:
		return ErrInvalidTokenError
	}

	return nil
}

type Null struct {
	TerminalNodeBase
}

func (n *Null) GetPrimitiveType() typesystem.PrimitiveKind {
	return typesystem.PrimitiveKindInterface
}

func (n *Null) GetValue() any { return nil }

func (n *Null) SetTerminalToken(tk IToken) error {
	n.Start = tk
	n.End = tk

	return nil
}
