package psi

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
)

type LanguageID string

type Language interface {
	Name() LanguageID
	Extensions() []string

	CreateSourceFile(fileName string, fileHandle repofs.FileHandle) SourceFile

	Parse(fileName string, code string) (SourceFile, error)
	ParseCodeBlock(name string, block mdutils.CodeBlock) (SourceFile, error)
}

type SourceFile interface {
	Node

	Name() string
	Language() Language

	Root() Node
	Error() error

	Load() error
	Replace(code string) error

	OriginalText() string
	ToCode(node Node) (mdutils.CodeBlock, error)

	MergeCompletionResults(ctx context.Context, scope Scope, cursor Cursor, newSource SourceFile, newAst Node) error
}

type Scope interface {
	Root() Node
}
