package project

import (
	"context"
	"go/token"

	"github.com/greenboxal/aip/aip-langchain/pkg/llm"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/fti"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphstore"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/tasks"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/langs"
)

type Project interface {
	psi.Node

	TaskManager() *tasks.Manager
	LogManager() *thoughtdb.Repo

	RootPath() string
	RootNode() psi.Node
	Repo() *fti.Repository
	VcsFileSystem() repofs.FS
	FileSet() *token.FileSet
	Graph() *graphstore.IndexedGraph
	LanguageProvider() *Registry
	Embedder() llm.Embedder

	Sync(ctx context.Context) error
	Commit() error

	GetSourceFile(ctx context.Context, path string) (langs.SourceFile, error)
}
