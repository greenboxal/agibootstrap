package restv1

import (
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/indexing"
)

type SearchResponse struct {
	psi.NodeBase

	Results []indexing.NodeSearchHit `json:"results"`
}

var SearchResponseType = psi.DefineNodeType[*SearchResponse]()
