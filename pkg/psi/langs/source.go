package langs

import (
	"context"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
)

type SourceFile interface {
	psi.Node

	Name() string
	Language() Language

	Root() psi.Node
	Error() error

	Load(ctx context.Context) error
	Replace(ctx context.Context, code string) error

	OriginalText() string
	ToCode(node psi.Node) (mdutils.CodeBlock, error)

	MergeCompletionResults(ctx context.Context, scope Scope, cursor psi.Cursor, newSource SourceFile, newAst psi.Node) error
}

type Scope interface {
	Root() psi.Node
}

type SourceFileBase struct {
	psi.NodeBase
}

type FileHandleSource interface {
	psi.Node

	Open() (repofs.FileHandle, error)
}

func (sfb *SourceFileBase) GetFileHandle() (repofs.FileHandle, error) {
	parent, ok := sfb.Parent().(FileHandleSource)

	if !ok {
		return nil, errors.New("parent is not a FileHandleSource")
	}

	return parent.Open()
}
