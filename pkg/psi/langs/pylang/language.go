package pylang

import (
	"bytes"
	"context"
	"io"

	"github.com/pkg/errors"

	project2 "github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/psi/langs"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
)

const LanguageID langs.LanguageID = "python"

func init() {
	project2.RegisterLanguage(LanguageID, func(p project2.Project) langs.Language {
		return NewLanguage(p)
	})
}

type Language struct {
	project project2.Project
}

func NewLanguage(p project2.Project) *Language {
	return &Language{
		project: p,
	}
}

func (l *Language) Name() langs.LanguageID {
	return LanguageID
}

func (l *Language) Extensions() []string {
	return []string{".py"}
}

func (l *Language) CreateSourceFile(ctx context.Context, fileName string, fileHandle repofs.FileHandle) langs.SourceFile {
	return NewSourceFile(l, fileName, fileHandle)
}

func (l *Language) Parse(ctx context.Context, fileName string, code string) (langs.SourceFile, error) {
	f := l.CreateSourceFile(ctx, fileName, &BufferFileHandle{data: code})

	if err := f.Load(ctx); err != nil {
		return nil, err
	}

	return f, nil
}

func (l *Language) ParseCodeBlock(ctx context.Context, blockName string, block mdutils.CodeBlock) (langs.SourceFile, error) {
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
