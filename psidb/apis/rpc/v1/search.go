package rpcv1

import (
	"context"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/greenboxal/agibootstrap/psidb/core"
)

type Search struct {
	core *core.Core
}

func NewSearch(core *core.Core) *Search {
	return &Search{core: core}
}

type SearchRequest struct {
	Link cidlink.Link `json:"link"`
	Raw  bool         `json:"raw"`
}

type SearchResponse struct {
	Data []byte `json:"data"`
}

func (s *Search) Search(ctx context.Context, req *SearchRequest) (res *SearchResponse, err error) {

	return
}
