package fti

import (
	"os"
	"path/filepath"
)

type Repository struct {
}

func NewRepository(rootPath string) (*Repository, error) {
	r := &Repository{
		repoPath: rootPath,
		ftiPath:  filepath.Join(rootPath, ".fti"),
	}

	if err := r.loadConfig(); err != nil {
		if err != ErrNoConfig {
			return nil, err
		}
	}

	if err := r.loadIgnoreFile(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	r.index, err = NewOnlineIndex(r)

	if err != nil {
		return nil, err
	}

	if err := r.loadIndex(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return r, nil
}
