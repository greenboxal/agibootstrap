package iam

import "github.com/greenboxal/agibootstrap/psidb/psi"

type Identity struct {
	psi.NodeBase

	Username string `json:"username"`
}

var IdentityType = psi.DefineNodeType[*Identity]()

func (i *Identity) PsiNodeName() string { return i.Username }
