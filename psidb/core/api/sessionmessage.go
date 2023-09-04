package coreapi

import "github.com/greenboxal/agibootstrap/pkg/typesystem"

type SessionMessage interface {
	SessionMessageMarker()
	SessionMessageHeader() SessionMessageHeader
	SetSessionMessageHeader(header SessionMessageHeader)
}

type SessionMessageHeader struct {
	MessageID uint64 `json:"message_id"`
	ReplyToID uint64 `json:"reply_to_id"`
	SessionID string `json:"session_id"`
}

type SessionMessageBase struct {
	MessageHeader SessionMessageHeader `json:"-"`
}

func (sm *SessionMessageBase) SessionMessageMarker()                      {}
func (sm *SessionMessageBase) SessionMessageHeader() SessionMessageHeader { return sm.MessageHeader }
func (sm *SessionMessageBase) SetSessionMessageHeader(header SessionMessageHeader) {
	sm.MessageHeader = header
}

type SessionMessageOpen struct {
	SessionMessageBase

	Config SessionConfig `json:"config"`
}

type SessionMessageAck struct {
	SessionMessageBase
}

type SessionMessageNack struct {
	SessionMessageBase
}

type SessionMessageKeepAlive struct {
	SessionMessageBase

	Timestamp int64 `json:"timestamp"`
}

type SessionMessageShutdown struct {
	SessionMessageBase
}

func init() {
	typesystem.GetType[SessionMessage]()
	typesystem.GetType[SessionMessageOpen]()
	typesystem.GetType[SessionMessageAck]()
	typesystem.GetType[SessionMessageNack]()
	typesystem.GetType[SessionMessageShutdown]()
}
