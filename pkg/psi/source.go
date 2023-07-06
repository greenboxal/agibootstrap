package psi

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
)

type LanguageID string

type Language interface {
	Name() LanguageID
	Extensions() []string

	CreateSourceFile(ctx context.Context, fileName string, fileHandle repofs.FileHandle) SourceFile

	Parse(ctx context.Context, fileName string, code string) (SourceFile, error)
	ParseCodeBlock(ctx context.Context, name string, block mdutils.CodeBlock) (SourceFile, error)
}

type SourceFile interface {
	Node

	Name() string
	Language() Language

	Root() Node
	Error() error

	Load(ctx context.Context) error
	Replace(ctx context.Context, code string) error

	OriginalText() string
	ToCode(node Node) (mdutils.CodeBlock, error)

	MergeCompletionResults(ctx context.Context, scope Scope, cursor Cursor, newSource SourceFile, newAst Node) error
}

type Scope interface {
	Root() Node
}
