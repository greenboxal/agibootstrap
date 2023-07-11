package psi

import (
	"context"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
)

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

type AstNode interface {
	Node
}

type Scope interface {
	Root() Node
}

type SourceFileBase struct {
	NodeBase
}

type FileHandleSource interface {
	Node

	Open() (repofs.FileHandle, error)
}

func (sfb *SourceFileBase) GetFileHandle() (repofs.FileHandle, error) {
	parent, ok := sfb.Parent().(FileHandleSource)

	if !ok {
		return nil, errors.New("parent is not a FileHandleSource")
	}

	return parent.Open()
}
