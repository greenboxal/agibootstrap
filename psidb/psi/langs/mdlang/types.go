package mdlang

import (
	"github.com/gomarkdown/markdown/ast"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

func init() {
	psi.DefineNodeType[*SourceFile]()
	psi.DefineNodeType[*NodeBase[ast.Node]]()
}
