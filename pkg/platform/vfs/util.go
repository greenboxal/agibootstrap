package vfs

import (
	"path/filepath"
	"strings"
)

func isChildPath(parent, child string) (bool, error) {
	// Clean the paths
	parent, err := filepath.Abs(parent)
	if err != nil {
		return false, err
	}
	child, err = filepath.Abs(child)
	if err != nil {
		return false, err
	}

	// Check the relation
	return strings.HasPrefix(child, parent), nil
}
