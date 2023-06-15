package fti

import (
	"os"
	"path/filepath"
)

type Repository struct {
}

func NewRepository(repoPath string) (*Repository, error) {
	r := &Repository{
		repoPath: repoPath,
		ftiPath:  filepath.Join(repoPath, ".fti"),
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

	var err error
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
