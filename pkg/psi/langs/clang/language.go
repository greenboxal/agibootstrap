package clang

import (
	"bytes"
	"context"
	"io"

	"github.com/pkg/errors"

	project2 "github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
)

const LanguageID project2.LanguageID = "c"

func init() {
	project2.RegisterLanguage(LanguageID, NewLanguage)
}

type Language struct {
	project project2.Project
}

func NewLanguage(p project2.Project) project2.Language {
	return &Language{
		project: p,
	}
}

func (l *Language) Name() project2.LanguageID {
	return LanguageID
}

func (l *Language) Extensions() []string {
	return []string{".c", ".h"}
}

func (l *Language) CreateSourceFile(ctx context.Context, fileName string, fileHandle repofs.FileHandle) project2.SourceFile {
	return NewSourceFile(l, fileName, fileHandle)
}

func (l *Language) Parse(ctx context.Context, fileName string, code string) (project2.SourceFile, error) {
	f := l.CreateSourceFile(ctx, fileName, &BufferFileHandle{data: code})

	if err := f.Load(ctx); err != nil {
		return nil, err
	}

	return f, nil
}

func (l *Language) ParseCodeBlock(ctx context.Context, blockName string, block mdutils.CodeBlock) (project2.SourceFile, error) {
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
