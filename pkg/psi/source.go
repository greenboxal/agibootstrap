package psi

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/repofs"
)

type LanguageID string

type Language interface {
	Name() LanguageID
	Extensions() []string

	CreateSourceFile(fileName string, fileHandle repofs.FileHandle) SourceFile

	Parse(fileName string, code string) (SourceFile, error)
	ParseCodeBlock(name string, block CodeBlock) (SourceFile, error)
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
	ToCode(node Node) (string, error)

	MergeCompletionResults(ctx context.Context, scope any, cursor Cursor, newAst Node) error
}

// CodeBlock represents a block of code with its language and code content.
type CodeBlock struct {
	Language string
	Code     string
}
