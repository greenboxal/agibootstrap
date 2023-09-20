package markdown

import "github.com/greenboxal/agibootstrap/psidb/utils/sparsing"

type Line struct {
	sparsing.TerminalNodeBase
}

type Document struct {
	sparsing.CompositeNodeBase
}
