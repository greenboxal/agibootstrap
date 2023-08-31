package psiml

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi `github.com/greenboxal/agibootstrap/psidb/core/api`
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
)

type Search struct {
	psi.NodeBase

	From  psi.Path  `json:"from"`
	Query QueryTerm `json:"query"`
	Limit int       `json:"limit"`

	View string `json:"view"`
}

type QueryTerm struct {
	psi.NodeBase

	Text string   `json:"text"`
	Node psi.Node `json:"node"`
	Path psi.Path `json:"path"`
}

var SearchType = psi.DefineNodeType[*Search]()
var QueryTermType = psi.DefineNodeType[*QueryTerm]()

func (q *QueryTerm) Resolve(ctx context.Context, lg coreapi.LiveGraph) (psi.Node, error) {
	if q.Node != nil {
		return q.Node, nil
	}

	if !q.Path.IsEmpty() {
		return lg.Resolve(ctx, q.Path)
	}

	return stdlib.NewText(q.Text), nil
}
