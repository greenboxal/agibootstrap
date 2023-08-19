package query

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type QueryContext interface {
	Context() context.Context
	Transaction() coreapi.Transaction
	Graph() psi.Graph
}

type Query interface {
	Run(ctx QueryContext, in Iterator) (Iterator, error)
}
