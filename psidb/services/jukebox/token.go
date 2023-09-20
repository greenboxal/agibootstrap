package jukebox

import "github.com/greenboxal/agibootstrap/psidb/utils/sparsing"

const (
	TokenKindInvalid sparsing.TokenKind = iota
	TokenKindIdentifier
	TokenKindLine
	TokenKindString
	TokenKindNumber
)
