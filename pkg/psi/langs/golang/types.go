package golang

import (
	"github.com/dave/dst"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func init() {
	psi.DefineNodeType[*SourceFile]()
	psi.DefineNodeType[*NodeBase[dst.Node]]()
	psi.DefineNodeType[*Container]()
	psi.DefineNodeType[*Leaf]()
}
