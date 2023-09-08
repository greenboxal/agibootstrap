package project

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type LanguageID string

func (l LanguageID) Name() string   { return string(l) }
func (l LanguageID) String() string { return l.Name() }

type Language interface {
	Name() LanguageID
	Extensions() []string

	CreateSourceFile(ctx context.Context, fileName string, fileHandle repofs.FileHandle) SourceFile

	Parse(ctx context.Context, fileName string, code string) (SourceFile, error)
	ParseCodeBlock(ctx context.Context, name string, block mdutils.CodeBlock) (SourceFile, error)
}

type LanguageBase struct {
}

func (l *LanguageBase) Extensions() []string {
	return nil
}

func (l *LanguageBase) Parse(ctx context.Context, fileName string, code string) (SourceFile, error) {
	//TODO implement me
	panic("implement me")
}

func (l *LanguageBase) ParseCodeBlock(ctx context.Context, name string, block mdutils.CodeBlock) (SourceFile, error) {
	//TODO implement me
	panic("implement me")
}

func (l *LanguageBase) CollectReferences(ctx context.Context, root psi.Node) error {
	//TODO implement me
	panic("implement me")
}
