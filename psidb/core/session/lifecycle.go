package session

import (
	"context"
	`time`

	`github.com/go-errors/errors`
	`github.com/jbenet/goprocess`
	goprocessctx `github.com/jbenet/goprocess/context`

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

func (sess *Session) Run(proc goprocess.Process) {
	proc.SetTeardown(sess.teardown)

	ctx := goprocessctx.OnClosingContext(proc)
	ctx = coreapi.WithSession(ctx, sess)

	defer func() {
		if e := recover(); e != nil {
			err := errors.Wrap(e, 1)

			sess.logger.Error(err)

			panic(err)
		}
	}()

	if err := sess.initialize(ctx); err != nil {
		panic(err)
	}

	if err := sess.virtualGraph.Recover(ctx); err != nil {
		panic(err)
	}

	if sess.config.KeepAliveTimeout == 0 {
		sess.config.KeepAliveTimeout = 30 * time.Second
		sess.config.Persistent = true
	}

	ticker := time.NewTicker(sess.config.KeepAliveTimeout)

	if !sess.config.Deadline.IsZero() {
		remaining := sess.config.Deadline.Sub(time.Now())

		time.AfterFunc(remaining, func() {
			sess.RequestShutdown()
		})
	}

	if sess.parent != nil {
		sess.parent.ReceiveMessage(&sessionMessageChildForked{child: sess})
	}

	sess.manager.onSessionStarted(sess)

	sess.status = SessionStateActive
	close(sess.readyCh)

	for {
		select {
		case _, ok := <-sess.stopCh:
			if !ok {
				return
			}

			if sess.TryShutdownNow() {
				return
			}

		case <-ticker.C:
			if !sess.config.Persistent {
				if time.Now().Sub(sess.lastKeepAlive) > 30*time.Second {
					sess.RequestShutdown()
				}
			}

		case msg := <-sess.IncomingMessageCh:
			if err := sess.processMessage(ctx, msg); err != nil {
				sess.logger.Error(err)
			}
		}
	}
}

func (sess *Session) TryShutdownNow() bool {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	if len(sess.children) > 0 {
		for _, child := range sess.children {
			child.ReceiveMessage(&coreapi.SessionMessageShutdown{})
		}

		return false
	}

	sess.ShutdownNow()

	return true
}

func (sess *Session) RequestShutdown() {
	sess.stopCh <- struct{}{}
}

func (sess *Session) ShutdownNow() {
	if sess.status == SessionStateClosing || sess.status == SessionStateClosed {
		return
	}

	sess.status = SessionStateClosing

	if sess.stopCh != nil {
		close(sess.stopCh)
		sess.stopCh = nil
	}
}

func (sess *Session) ShutdownAndWait(ctx context.Context) error {
	sess.RequestShutdown()

	select {
	case _, _ = <-sess.proc.Closed():
		return sess.proc.Err()

	case <-ctx.Done():
		return ctx.Err()
	}
}

func (sess *Session) Close() error {
	sess.RequestShutdown()

	return sess.proc.Err()
}

func (sess *Session) teardown() error {
	ctx := context.Background()

	if err := sess.serviceProvider.Close(ctx); err != nil {
		return nil
	}

	if err := sess.ClientBusConnection.Close(); err != nil {
		return err
	}

	sess.status = SessionStateClosed

	if sess.parent != nil {
		sess.parent.ReceiveMessage(&sessionMessageChildFinished{child: sess})
	}

	sess.manager.onSessionFinish(sess)

	return nil
}
