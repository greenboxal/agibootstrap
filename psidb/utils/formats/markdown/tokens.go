package markdown

import "github.com/greenboxal/agibootstrap/psidb/utils/sparsing"

type TokenKind = sparsing.TokenKind
type Position = sparsing.Position
type Token = sparsing.Token

const (
	TokenKindInvalid TokenKind = iota
	TokenKindText
	TokenKindSymbol
	TokenKindSpace
	TokenKindLineBreak

	TokenKindEOS = sparsing.TokenKindEOS
	TokenKindEOF = sparsing.TokenKindEOF
)
