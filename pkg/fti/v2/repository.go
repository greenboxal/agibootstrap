package fti

import "context"

type Repository struct {
}

func NewRepository(rootPath string) (*Repository, error) {
	// TODO: Refactor the old Repository implementation into this file
	return &Repository{}, nil
}

func (r *Repository) Update(context.Context) error {
	// TODO: Implement update logic
	return nil
}

func (r *Repository) Init() error {
	// TODO: Implement initialization logic
	return nil
}
