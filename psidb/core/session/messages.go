package session

type sessionMessageChildFinished struct {
	child *Session
}

type sessionMessageChildForked struct {
	child *Session
}

func (s sessionMessageChildFinished) SessionMessageMarker() {}
func (s sessionMessageChildForked) SessionMessageMarker()   {}
