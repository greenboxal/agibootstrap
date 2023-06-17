package mdlang

import (
	"bytes"
	"io"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/codex"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
)

const LanguageID psi.LanguageID = "markdown"

func init() {
	codex.RegisterLanguage(LanguageID, NewLanguage)
}

type Language struct {
	project *codex.Project
}

func NewLanguage(p *codex.Project) psi.Language {
	return &Language{
		project: p,
	}
}

func (l *Language) Name() psi.LanguageID {
	return LanguageID
}

func (l *Language) Extensions() []string {
	return []string{".md"}
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

func (l *Language) ParseCodeBlock(blockName string, block psi.CodeBlock) (psi.SourceFile, error) {
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
