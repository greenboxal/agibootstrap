package session

import coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"

type sessionMessageChildFinished struct {
	coreapi.SessionMessageBase

	child *Session
}

type sessionMessageChildForked struct {
	coreapi.SessionMessageBase

	child *Session
}
