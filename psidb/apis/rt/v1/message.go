package rtv1

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"
)

type MessageFlags uint8

const (
	MessageFlagNone            MessageFlags = 0x00
	MessageFlagHasContinuation              = 0x01
)

type MessageKind uint8

const (
	MessageKindNop MessageKind = iota

	MessageKindAck
	MessageKindNack

	MessageKindLookupNode
	MessageKindReadNode
	MessageKindReadEdge
	MessageKindReadEdges
	MessageKindPushFrame
)

type MessageHeader struct {
	Kind      MessageKind  `json:"kind"`
	Flags     MessageFlags `json:"flags"`
	MessageID uint64       `json:"mid"`
	ReplyToID uint64       `json:"rtid"`
}

func (h *MessageHeader) GetMessageID() uint64             { return h.MessageID }
func (h *MessageHeader) GetMessageKind() MessageKind      { return h.Kind }
func (h *MessageHeader) GetMessageHeader() *MessageHeader { return h }

type Message interface {
	GetMessageID() uint64
	GetMessageKind() MessageKind
	GetMessageHeader() *MessageHeader
}

func DefineMessageReplyWithType[TReq Message, TResp Message](kind MessageKind) MessageWithReplyType[TReq, TResp] {
	return MessageWithReplyType[TReq, TResp]{
		Kind:         kind,
		RequestType:  typesystem.GetType[TReq](),
		ResponseType: typesystem.GetType[TResp](),
	}
}

type MessageWithReplyType[TReq Message, TResp Message] struct {
	Kind         MessageKind
	RequestType  typesystem.Type
	ResponseType typesystem.Type
}

func (m MessageWithReplyType[TReq, TResp]) BuildRequest(client *Client, req TReq) Message {
	hdr := req.GetMessageHeader()
	hdr.Kind = m.Kind
	hdr.MessageID = client.getNextMessageID()

	return req
}

func (m MessageWithReplyType[TReq, TResp]) MakeCall(ctx context.Context, client *Client, req TReq) (empty TResp, _ error) {
	msg := m.BuildRequest(client, req)
	ch := client.tracker.AddPendingCall(msg.GetMessageID(), 1, false)

	if err := client.SendMessage(ctx, msg); err != nil {
		return empty, err
	}

	select {
	case <-ctx.Done():
		return empty, ctx.Err()
	case resp := <-ch:
		if resp.Error != nil {
			return empty, resp.Error
		} else {
			return resp.Result.(TResp), nil
		}
	}
}

func (m MessageWithReplyType[TReq, TResp]) MakeCallStreamed(ctx context.Context, client *Client, req TReq, bufferHint int) (<-chan RpcResponse, error) {
	if bufferHint <= 0 {
		bufferHint = 1
	}

	msg := m.BuildRequest(client, req)
	ch := client.tracker.AddPendingCall(msg.GetMessageID(), bufferHint, false)

	if err := client.SendMessage(ctx, msg); err != nil {
		return nil, err
	}

	return ch, nil
}
