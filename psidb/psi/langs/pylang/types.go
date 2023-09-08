package pylang

import (
	"github.com/antlr4-go/antlr/v4"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

func init() {
	psi.DefineNodeType[*SourceFile]()
	psi.DefineNodeType[*NodeBase[antlr.ParseTree]]()
}
