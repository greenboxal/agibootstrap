package mdlang

import (
	"github.com/gomarkdown/markdown/ast"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func init() {
	psi.DefineNodeType[*SourceFile]()
	psi.DefineNodeType[*NodeBase[ast.Node]]()
}
