package psi

import (
	"io/fs"

	"github.com/pkg/errors"
)

var ErrNodeNotFound = errors.Wrap(fs.ErrNotExist, "node not found")

var ErrAbort = errors.New("abort")
