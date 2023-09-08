package project

import (
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Library interface {
	Module
}

type LibraryManager struct {
	psi.NodeBase
}
