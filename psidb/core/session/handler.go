package session

import (
	"context"
	"time"

	"golang.org/x/exp/slices"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

func (sess *Session) processMessage(ctx context.Context, msg coreapi.SessionMessage) error {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	sess.lastKeepAlive = time.Now()

	switch msg := msg.(type) {
	case *coreapi.SessionMessageKeepAlive:
		return nil

	case *sessionMessageChildForked:
		for _, child := range sess.children {
			if child.UUID() == msg.child.UUID() {
				return nil
			}
		}

		sess.children = append(sess.children, msg.child)

		return nil

	case *sessionMessageChildFinished:
		idx := slices.Index(sess.children, msg.child)

		if idx != -1 {
			sess.children = slices.Delete(sess.children, idx, idx+1)
		}

		if sess.status == SessionStateClosing && len(sess.children) == 0 {
			sess.RequestShutdown()
		}

		return nil

	case *coreapi.SessionMessageShutdown:
		sess.RequestShutdown()

		return nil

	case *coreapi.SessionMessageOpen:
		sess.SendReply(msg, &coreapi.SessionMessageOpen{
			Config: sess.config,
		})

		return nil

	default:
		logger.Warn("unknown message type: %T", msg)

		return nil
	}
}
