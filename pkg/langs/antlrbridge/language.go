package antlrbridge

import (
	"github.com/greenboxal/agibootstrap/pkg/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/project"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
)

type Language struct {
	self    psi.Language
	project project.Project
}

func (l *Language) Init(self psi.Language, p project.Project) {
	l.project = p
	l.self = self
}

func (l *Language) CreateSourceFile(fileName string, fileHandle repofs.FileHandle) psi.SourceFile {
	return NewSourceFile(l, fileName, fileHandle)
}

func (l *Language) Parse(fileName string, code string) (psi.SourceFile, error) {
	f := l.CreateSourceFile(fileName, repofs.String(code))

	if err := f.Load(); err != nil {
		return nil, err
	}

	return f, nil
}

func (l *Language) ParseCodeBlock(blockName string, block mdutils.CodeBlock) (psi.SourceFile, error) {
	return l.Parse(blockName, block.Code)
}
