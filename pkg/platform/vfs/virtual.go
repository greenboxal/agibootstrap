package vfs

import (
	"context"
	"io"
	"io/fs"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type VirtualFileSystem struct {
	Root *Directory
}

func (v *VirtualFileSystem) Open(name string) (fs.File, error) {
	n, err := psi.Resolve(context.Background(), v.Root, name)

	if err != nil {
		return nil, err
	}

	return n.(fs.File), nil
}

func (v *VirtualFileSystem) WriteFile(p string, src io.Reader) error {
	//TODO implement me
	panic("implement me")
}
