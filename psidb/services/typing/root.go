package typing

import "github.com/greenboxal/agibootstrap/pkg/psi"

type Root struct {
	psi.NodeBase
}

var RootType = psi.DefineNodeType[*Root]()

func NewRoot() *Root {
	r := &Root{}
	r.Init(r)

	return r
}

func (r *Root) PsiNodeName() string { return "_VTS" }
