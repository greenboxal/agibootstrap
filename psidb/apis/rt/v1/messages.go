package rtv1

import (
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/client"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
)

var MessageTypeLookupNode = DefineMessageReplyWithType[*LookupNodeRequest, *LookupNodeResponse](MessageKindReadNode)
var MessageTypeReadNode = DefineMessageReplyWithType[*ReadNodeRequest, *ReadNodeResponse](MessageKindReadNode)
var MessageTypeReadEdge = DefineMessageReplyWithType[*ReadEdgeRequest, *ReadEdgeResponse](MessageKindReadEdge)
var MessageTypeReadEdges = DefineMessageReplyWithType[*ReadEdgesRequest, *ReadEdgesResponse](MessageKindReadEdges)
var MessageTypePushFrame = DefineMessageReplyWithType[*PushFrameRequest, *PushFrameResponse](MessageKindPushFrame)

type LookupNodeRequest struct {
	MessageHeader

	Inode int64    `json:"inode"`
	Path  psi.Path `json:"path"`
}

type LookupNodeResponse struct {
	MessageHeader

	Data *coreapi.SerializedNode `json:"data"`
}

type ReadNodeRequest struct {
	MessageHeader

	Inode int64    `json:"inode"`
	Path  psi.Path `json:"path"`
}

type ReadNodeResponse struct {
	MessageHeader

	Data *coreapi.SerializedNode `json:"data"`
}

type ResolveRequest struct {
	MessageHeader

	Path psi.Path `json:"path"`
}

type ReadEdgeRequest struct {
	MessageHeader

	Path        psi.Path `json:"path"`
	ParentInode int64    `json:"parent_inode"`
}

type ReadEdgeResponse struct {
	MessageHeader

	Data *coreapi.SerializedEdge `json:"data"`
}

type ReadEdgesRequest struct {
	MessageHeader

	Inode int64    `json:"inode"`
	Path  psi.Path `json:"path"`
}

type ReadEdgesResponse struct {
	MessageHeader

	Data []*coreapi.SerializedEdge `json:"data"`
}

type PushFrameRequest struct {
	MessageHeader

	Frame *client.Frame `json:"frame"`
}

type PushFrameResponse struct {
	MessageHeader
}
