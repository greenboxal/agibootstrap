package codex

import (
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type analysisListener struct {
	p *Project
}

func (a *analysisListener) OnNodeUpdated(node psi.Node) {

}
