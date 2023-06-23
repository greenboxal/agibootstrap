package project

import (
	"go/token"

	"github.com/greenboxal/agibootstrap/pkg/fti"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
	"github.com/greenboxal/agibootstrap/pkg/tasks"
)

type Project interface {
	psi.Node

	TaskManager() *tasks.Manager

	RootPath() string
	RootNode() psi.Node
	Repo() *fti.Repository
	FS() repofs.FS
	FileSet() *token.FileSet
	Graph() psi.Graph
	LanguageProvider() *Registry

	Sync() error
	Commit() error

	GetSourceFile(path string) (psi.SourceFile, error)
}
