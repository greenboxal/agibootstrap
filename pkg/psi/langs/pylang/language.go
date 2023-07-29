package pylang

import (
	"bytes"
	"context"
	"io"

	"github.com/pkg/errors"

	project "github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
)

const LanguageID project.LanguageID = "python"

func init() {
	project.RegisterLanguage(LanguageID, func(p project.Project) project.Language {
		return NewLanguage(p)
	})
}

type Language struct {
	project.LanguageBase

	project project.Project
}

func NewLanguage(p project.Project) *Language {
	return &Language{
		project: p,
	}
}

func (l *Language) Name() project.LanguageID {
	return LanguageID
}

func (l *Language) Extensions() []string {
	return []string{".py"}
}

func (l *Language) CreateSourceFile(ctx context.Context, fileName string, fileHandle repofs.FileHandle) project.SourceFile {
	return NewSourceFile(l, fileName, fileHandle)
}

func (l *Language) Parse(ctx context.Context, fileName string, code string) (project.SourceFile, error) {
	f := l.CreateSourceFile(ctx, fileName, &BufferFileHandle{data: code})

	if err := f.Load(ctx); err != nil {
		return nil, err
	}

	return f, nil
}

func (l *Language) ParseCodeBlock(ctx context.Context, blockName string, block mdutils.CodeBlock) (project.SourceFile, error) {
	return l.Parse(ctx, blockName, block.Code)
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
