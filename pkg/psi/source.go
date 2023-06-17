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
	ToCode(node Node) (string, error)

	MergeCompletionResults(ctx context.Context, scope Scope, cursor Cursor, newAst Node) error
}

type Scope interface {
	Root() Node
}

func AsCodeBlock(sf SourceFile, node Node) (mdutils.CodeBlock, error) {
	code, err := sf.ToCode(node)

	if err != nil {
		return mdutils.CodeBlock{}, err
	}

	return mdutils.CodeBlock{
		Filename: sf.Name(),
		Language: string(sf.Language().Name()),
		Code:     code,
	}, nil
}
