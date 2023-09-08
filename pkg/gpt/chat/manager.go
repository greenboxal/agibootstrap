package chat

import (
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Manager struct {
	psi.NodeBase
}

func NewManager() *Manager {
	return &Manager{}
}
