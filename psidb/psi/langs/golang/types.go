package golang

import (
	"github.com/dave/dst"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

func init() {
	psi.DefineNodeType[*SourceFile]()
	psi.DefineNodeType[*NodeBase[dst.Node]]()
	psi.DefineNodeType[*Container]()
	psi.DefineNodeType[*Leaf]()
}
