package client

import "sync"

type RpcResponse struct {
	MessageID uint64
	Error     error
	Result    Message
}

type pendingRpcCall struct {
	mid    uint64
	ch     chan RpcResponse
	single bool
}

type RpcResponseTracker struct {
	mu           sync.RWMutex
	pendingCalls map[uint64]*pendingRpcCall
}

func NewRpcResponseTracker() *RpcResponseTracker {
	return &RpcResponseTracker{
		pendingCalls: map[uint64]*pendingRpcCall{},
	}
}

func (t *RpcResponseTracker) AddPendingCall(mid uint64, bufferLen int, isMultiResponse bool) chan RpcResponse {
	t.mu.Lock()
	defer t.mu.Unlock()

	if bufferLen <= 0 {
		bufferLen = 1
	}

	pending := &pendingRpcCall{
		mid:    mid,
		ch:     make(chan RpcResponse, bufferLen),
		single: !isMultiResponse,
	}

	t.pendingCalls[mid] = pending

	return pending.ch
}

func (t *RpcResponseTracker) AcceptMessage(msg Message) {
	hdr := msg.GetMessageHeader()
	replyTo := hdr.ReplyToID

	if replyTo == 0 {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	call := t.pendingCalls[replyTo]

	if call != nil {
		return
	}

	call.ch <- RpcResponse{
		MessageID: replyTo,
		Error:     nil,
		Result:    msg,
	}

	if call.single || (hdr.Flags&MessageFlagHasContinuation) == 0 {
		delete(t.pendingCalls, replyTo)
	}
}
