package coreapi

import "github.com/greenboxal/agibootstrap/pkg/typesystem"

type SessionMessage interface {
	SessionMessageMarker()
}

type SessionMessageOpen struct {
}

type SessionMessageAck struct {
}

type SessionMessageKeepAlive struct {
	Timestamp int64 `json:"timestamp"`
}

type SessionMessageShutdown struct {
}

func (sm SessionMessageOpen) SessionMessageMarker()      {}
func (sm SessionMessageAck) SessionMessageMarker()       {}
func (sm SessionMessageKeepAlive) SessionMessageMarker() {}
func (sm SessionMessageShutdown) SessionMessageMarker()  {}

func init() {
	typesystem.GetType[SessionMessage]()
	typesystem.GetType[SessionMessageOpen]()
	typesystem.GetType[SessionMessageAck]()
	typesystem.GetType[SessionMessageShutdown]()
}
