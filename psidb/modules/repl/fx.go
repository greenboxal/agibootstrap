package repl

import (
	"context"
	"net"
	"os"
	"path"
	"strings"

	"go.uber.org/fx"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

var Module = fx.Module(
	"modules/repl",

	fx.Provide(NewManager),

	fx.Invoke(func(m *Manager) {}),
)

type Manager struct {
	sm coreapi.SessionManager
}

func NewManager(
	lc fx.Lifecycle,
	core coreapi.Core,
	sm coreapi.SessionManager,
) *Manager {
	m := &Manager{
		sm: sm,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return m.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return m.Stop(ctx)
		},
	})
	return m
}

func (m *Manager) Start(ctx context.Context) error {
	repl0Path := "/tmp/psidb/repl0"

	if err := os.MkdirAll(path.Dir(repl0Path), 0777); err != nil {
		return err
	}

	_ = os.Remove(repl0Path)

	socket, err := net.Listen("unix", repl0Path)

	if err != nil {
		return err
	}

	startStream := func(fd net.Conn) {
		sess := m.sm.GetOrCreateSession(coreapi.SessionConfig{})
		instance := NewServerReplHandler(ctx, sess, StreamStandardIO{
			Out: fd,
			Err: fd,
			In:  fd,
		})

		var buffer string

		for {
			buf := make([]byte, 1024)
			n, err := fd.Read(buf)

			if err != nil {
				return
			}

			buffer += string(buf[:n])
			lfIndex := strings.Index(buffer, "\n")

			if lfIndex != -1 {
				line := buffer[:lfIndex]
				buffer = buffer[lfIndex+1:]

				result, err := instance.RunLine(ctx, line)

				if err != nil {
					_, _ = fd.Write([]byte(err.Error()))
				} else {
					_, _ = fd.Write([]byte(result.(string)))
				}

				_, _ = fd.Write([]byte("\n"))
			}
		}
	}

	go func() {
		for {
			client, err := socket.Accept()

			if err != nil {
				panic(err)
			}

			go startStream(client)
		}
	}()

	return nil
}

func (m *Manager) Stop(ctx context.Context) error {
	return nil
}
