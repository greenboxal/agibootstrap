package rest

import (
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/db/indexing"
)

type SearchResponse struct {
	psi.NodeBase

	Results []indexing.NodeSearchHit `json:"results"`
}

var SearchResponseType = psi.DefineNodeType[*SearchResponse]()
