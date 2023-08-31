package rtv1

import (
	"context"

	"github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

type CreateRequest struct {
	Parent int64                 `json:"parent"`
	Key    string                `json:"key"`
	Flags  graphfs.OpenNodeFlags `json:"flags"`
}

type CreateResponse struct {
	Handle int64 `json:"handle"`
}

type LookupRequest struct {
}

type LookupResponse struct {
	Handle int64 `json:"handle"`
}

type ReadRequest struct {
	Handle int64 `json:"handle"`
}

type ReadResponse struct {
	Node *coreapi.SerializedNode `json:"node"`
}

type WriteRequest struct {
	Handle int64                   `json:"handle"`
	Node   *coreapi.SerializedNode `json:"node"`
}

type WriteResponse struct {
}

type SetEdgeRequest struct {
	Handle int64                   `json:"handle"`
	Edge   *coreapi.SerializedEdge `json:"edge"`
}

type SetEdgeResponse struct {
}

type RemoveEdgeRequest struct {
	Handle int64                   `json:"handle"`
	Edge   *coreapi.SerializedEdge `json:"edge"`
}

type RemoveEdgeResponse struct {
}

type ReadEdgeRequest struct {
	Handle int64                   `json:"handle"`
	Edge   *coreapi.SerializedEdge `json:"edge"`
}

type ReadEdgeResponse struct {
	Edge *coreapi.SerializedEdge `json:"edge"`
}

type ReadEdgesRequest struct {
	Handle int64                   `json:"handle"`
	Edge   *coreapi.SerializedEdge `json:"edge"`
}

type ReadEdgesResponse struct {
	Edges []*coreapi.SerializedEdge `json:"edges"`
}

type NodeAPI interface {
	Create(ctx context.Context, request *CreateRequest) (*CreateResponse, error)
	Lookup(ctx context.Context, request *LookupRequest) (*LookupResponse, error)

	Read(ctx context.Context, request *WriteRequest) (*ReadResponse, error)
	Write(ctx context.Context, request *WriteRequest) (*WriteResponse, error)

	SetEdge(ctx context.Context, request *SetEdgeRequest) (*SetEdgeResponse, error)
	RemoveEdge(ctx context.Context, request *RemoveEdgeRequest) (*RemoveEdgeResponse, error)

	ReadEdge(ctx context.Context, request *ReadEdgeRequest) (*ReadEdgeResponse, error)
	ReadEdges(ctx context.Context, request *ReadEdgesRequest) (*ReadEdgesResponse, error)
}
