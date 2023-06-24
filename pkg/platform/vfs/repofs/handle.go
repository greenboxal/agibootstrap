package repofs

import (
	"bytes"
	"io"
	"io/fs"
	"os"
)

type FileHandle interface {
	Get() (io.ReadCloser, error)
	Put(src io.Reader) error
	Close() error
}

type memoryFileHandle struct {
	content []byte
}

func (s *memoryFileHandle) Get() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewBuffer(s.content)), nil
}

func (s *memoryFileHandle) Put(src io.Reader) error {
	data, err := io.ReadAll(src)

	if err != nil {
		return err
	}

	s.content = data

	return nil
}

func (s *memoryFileHandle) Close() error {
	return nil
}

func String(content string) FileHandle {
	return Bytes([]byte(content))
}

func Bytes(content []byte) FileHandle {
	return &memoryFileHandle{content}
}

type FsFileHandle struct {
	FS   fs.FS
	Path string
}

func (o FsFileHandle) Get() (io.ReadCloser, error) {
	return o.FS.Open(o.Path)
}

func (o FsFileHandle) Put(src io.Reader) error {
	data, err := io.ReadAll(src)

	if err != nil {
		return err
	}

	return os.WriteFile(o.Path, data, 0644)
}

func (o FsFileHandle) Close() error {
	return nil
}
