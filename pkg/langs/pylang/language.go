package pylang

import (
	"bytes"
	"io"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/project"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
)

const LanguageID psi.LanguageID = "python"

func init() {
	project.RegisterLanguage(LanguageID, func(p project.Project) psi.Language {
		return NewLanguage(p)
	})
}

type Language struct {
	project project.Project
}

func NewLanguage(p project.Project) *Language {
	return &Language{
		project: p,
	}
}

func (l *Language) Name() psi.LanguageID {
	return LanguageID
}

func (l *Language) Extensions() []string {
	return []string{".py"}
}

func (l *Language) CreateSourceFile(fileName string, fileHandle repofs.FileHandle) psi.SourceFile {
	return NewSourceFile(l, fileName, fileHandle)
}

func (l *Language) Parse(fileName string, code string) (psi.SourceFile, error) {
	f := l.CreateSourceFile(fileName, &BufferFileHandle{data: code})

	if err := f.Load(); err != nil {
		return nil, err
	}

	return f, nil
}

func (l *Language) ParseCodeBlock(blockName string, block mdutils.CodeBlock) (psi.SourceFile, error) {
	return l.Parse(blockName, block.Code)
}

type BufferFileHandle struct {
	data string
}

type closerReader struct {
	io.Reader
}

func (c closerReader) Close() error {
	return nil
}

func (b BufferFileHandle) Get() (io.ReadCloser, error) {
	return closerReader{bytes.NewBufferString(b.data)}, nil
}

func (b BufferFileHandle) Put(src io.Reader) error {
	return errors.New("cannot put to buffer file handle")
}

func (b BufferFileHandle) Close() error {
	return nil
}
