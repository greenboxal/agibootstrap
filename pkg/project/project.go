package project

import (
	"go/token"

	"github.com/greenboxal/agibootstrap/pkg/fti"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
)

type Project interface {
	psi.Node

	RootPath() string
	RootNode() psi.Node
	FS() repofs.FS
	FileSet() *token.FileSet

	Repo() *fti.Repository
	LanguageProvider() *Registry

	Sync() error
	Commit() error

	GetSourceFile(path string) (psi.SourceFile, error)
}
