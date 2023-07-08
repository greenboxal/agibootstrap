package pylang

import (
	"github.com/antlr4-go/antlr/v4"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func init() {
	psi.DefineNodeType[*SourceFile]()
	psi.DefineNodeType[*NodeBase[antlr.ParseTree]]()
}
