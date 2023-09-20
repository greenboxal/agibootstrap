package jsonstream

import "github.com/greenboxal/agibootstrap/psidb/utils/sparsing"

type TokenKind = sparsing.TokenKind
type Position = sparsing.Position
type Token = sparsing.Token

const (
	TokenKindInvalid TokenKind = iota
	TokenKindOpenObject
	TokenKindCloseObject
	TokenKindOpenArray
	TokenKindCloseArray
	TokenKindColon
	TokenKindComma
	TokenKindString
	TokenKindNumber
	TokenKindIdent

	TokenKindEOS = sparsing.TokenKindEOS
	TokenKindEOF = sparsing.TokenKindEOF
)
