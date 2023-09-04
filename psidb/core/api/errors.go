package coreapi

import (
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

var ErrNotFound = psi.ErrNodeNotFound
var ErrTransactionClosed = errors.New("transaction closed")
var ErrUnsupportedOperation = errors.New("unsupported operation")
var ErrInvalidNodeType = errors.New("invalid node type")
