package project

import (
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Library interface {
	Module
}

type LibraryManager struct {
	psi.NodeBase
}
