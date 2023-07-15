package project

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/analysis"
)

type Module interface {
	psi.NamedNode
	analysis.ScopedNode

	GetName() string

	GetProject() Project
	GetRoot() Package
}

type ModuleBase struct {
	psi.NodeBase

	Name string `json:"name" psi-attr:""`
}

func (m *ModuleBase) PsiNodeName() string           { return m.GetName() }
func (m *ModuleBase) PsiNodeScope() *analysis.Scope { return analysis.GetEffectiveNodeScope(m) }

func (m *ModuleBase) GetName() string { return m.Name }

func (m *ModuleBase) GetProject() (result Project) {
	return *psi.LoadEdge(m, EdgeKindProject.Singleton(), &result)
}

func (m *ModuleBase) GetRoot() (result Package) {
	return *psi.LoadEdge(m, EdgeKindPackage.Named("Root"), &result)
}

type Package interface {
	psi.NamedNode
	analysis.ScopedNode

	GetName() string
	GetModule() Module

	GetDirectory() *vfs.Directory
}

const EdgeKindProject = psi.TypedEdgeKind[Project]("codex.project")
const EdgeKindModule = psi.TypedEdgeKind[Module]("codex.module")
const EdgeKindDirectory = psi.TypedEdgeKind[*vfs.Directory]("codex.directory")
const EdgeKindPackage = psi.TypedEdgeKind[Package]("codex.pkg")

type PackageBase struct {
	psi.NodeBase

	Name string `json:"name" psi-attr:""`
}

func (p *PackageBase) PsiNodeName() string           { return p.GetName() }
func (p *PackageBase) PsiNodeScope() *analysis.Scope { return analysis.GetEffectiveNodeScope(p) }

func (p *PackageBase) GetName() string { return p.Name }

func (p *PackageBase) GetModule() (result Module) {
	return *psi.LoadEdge(p, EdgeKindModule.Singleton(), &result)
}

func (p *PackageBase) GetProject() (result Project) {
	return *psi.LoadEdge(p, EdgeKindProject.Singleton(), &result)
}

func (p *PackageBase) GetDirectory() (result *vfs.Directory) {
	return *psi.LoadEdge(p, EdgeKindDirectory.Singleton(), &result)
}
