package rpcv1

import (
	"context"

	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/greenboxal/agibootstrap/psidb/core"
)

type ObjectStore struct {
	core *core.Core
}

func NewObjectStore(core *core.Core) *ObjectStore {
	return &ObjectStore{core: core}
}

type GetObjectRequest struct {
	Link cidlink.Link `json:"link"`
	Raw  bool         `json:"raw"`
}

type GetObjectResponse struct {
	Data []byte `json:"data"`
}

func (os *ObjectStore) GetObject(ctx context.Context, req *GetObjectRequest) (res *GetObjectResponse, err error) {
	lctx := linking.LinkContext{Ctx: ctx}
	n, err := os.core.LinkSystem().LoadRaw(lctx, req.Link)

	if err != nil {
		return
	}

	res = &GetObjectResponse{
		Data: n,
	}

	return
}
