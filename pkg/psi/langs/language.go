package langs

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
)

type LanguageID string

type Language interface {
	Name() LanguageID
	Extensions() []string

	CreateSourceFile(ctx context.Context, fileName string, fileHandle repofs.FileHandle) SourceFile

	Parse(ctx context.Context, fileName string, code string) (SourceFile, error)
	ParseCodeBlock(ctx context.Context, name string, block mdutils.CodeBlock) (SourceFile, error)
}
