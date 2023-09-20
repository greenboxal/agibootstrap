package jukebox

import (
	"context"
	"sync"

	"gitlab.com/gomidi/midi/v2"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
)

type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[uint8]*Connection
}

func NewConnectionManager(
	ctx inject.ResolutionContext,
) *ConnectionManager {
	cm := &ConnectionManager{
		connections: make(map[uint8]*Connection),
	}

	ctx.AppendShutdownHook(func(ctx context.Context) error {
		return cm.Close()
	})

	return cm
}

func (m *ConnectionManager) GetOrCreateConnection(channel uint8) *Connection {
	m.mu.RLock()
	if conn, ok := m.connections[channel]; ok {
		m.mu.RUnlock()
		return conn
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if conn, ok := m.connections[channel]; ok {
		return conn
	}

	out, err := midi.OutPort(0)

	if err != nil {
		panic(err)
	}

	conn := NewConnection(channel, out)

	m.connections[channel] = conn

	return conn
}

func (m *ConnectionManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, conn := range m.connections {
		if err := conn.Close(); err != nil {
			return err
		}
	}

	return nil
}
