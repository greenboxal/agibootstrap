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
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Project interface {
	psi.Node

	Graph() *graphstore.IndexedGraph

	TaskManager() *tasks.Manager
	LogManager() *thoughtdb.Repo

	Repo() *fti.Repository
	RootPath() string
	RootNode() psi.Node
	VcsFileSystem() repofs.FS

	FileTypeProvider() *FileTypeProvider
	LanguageProvider() *LanguageProvider
	Embedder() llm.Embedder

	Sync(ctx context.Context) error
	Commit() error

	GetSourceFile(ctx context.Context, path string) (SourceFile, error)

	FileSet() *token.FileSet
}
