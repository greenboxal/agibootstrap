package codex

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/langs"
)

type Module struct {
	project.ModuleBase

	p      *Project
	lang   langs.Language
	config project.ModuleConfig
}

var ModuleType = psi.DefineNodeType[*Module](psi.WithRuntimeOnly())

func NewModule(p *Project, cfg project.ModuleConfig, lang langs.Language, root *vfs.Directory) (*Module, error) {
	m := &Module{}
	m.p = p
	m.lang = lang
	m.config = cfg
	m.Name = cfg.Name
	m.Init(m, psi.WithNodeType(ModuleType))

	psi.UpdateEdge(m, project.EdgeKindProject.Singleton(), project.Project(p))

	return m, nil
}
