package golang

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
)

var LanguageID project.LanguageID = "go"

type Language struct{}

func (l *Language) Name() project.LanguageID { return LanguageID }
func (l *Language) Extensions() []string     { return []string{".go"} }

func (l *Language) CreateSourceFile(ctx context.Context, fileName string, fileHandle repofs.FileHandle) project.SourceFile {
	return NewSourceFile(l, fileName, fileHandle)
}

func (l *Language) Parse(ctx context.Context, fileName string, code string) (project.SourceFile, error) {
	sf := l.CreateSourceFile(ctx, fileName, repofs.String(code))

	if err := sf.Load(ctx); err != nil {
		return nil, err
	}

	return sf, nil
}

func (l *Language) ParseCodeBlock(ctx context.Context, name string, block mdutils.CodeBlock) (project.SourceFile, error) {
	return l.Parse(ctx, name, block.Code)
}

func (l *Language) OnEnabled(p project.Project) {
	ft := project.LanguageFileTypeBase{}
	ft.Name = "Go"
	ft.Language = l
	ft.Extensions = []string{".go"}

	p.FileTypeProvider().Register(&ft)
}

func init() {
	project.RegisterLanguage(LanguageID, func(p project.Project) project.Language {
		return &Language{}
	})
}
