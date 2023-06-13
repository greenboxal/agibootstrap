package repofs

import (
	"io"
	"io/fs"
	"os"
)

type FileHandle interface {
	Get() (fs.File, error)
	Put(src io.Reader) error
	Close() error
}

type FsFileHandle struct {
	FS   fs.FS
	Path string
}

func (o FsFileHandle) Get() (fs.File, error) {
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
