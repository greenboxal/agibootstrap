package query

import (
	"context"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type QueryContext interface {
	Context() context.Context
	Transaction() coreapi.Transaction
	Graph() psi.Graph
}

type Query interface {
	Run(ctx QueryContext, in Iterator) (Iterator, error)
}
