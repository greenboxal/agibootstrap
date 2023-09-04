package session

import (
	"sync"

	"github.com/jbenet/goprocess"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type BusConnection interface {
	ReceiveMessage(m coreapi.SessionMessage)
	SendMessage(m coreapi.SessionMessage)
}

type BusConnectionBase struct {
	IncomingMessageCh chan coreapi.SessionMessage
	OutgoingMessageCh chan coreapi.SessionMessage
}

func NewBusConnectionWithLimits(incoming, outgoing int) *BusConnectionBase {
	return NewBusConnectionWithChannels(
		make(chan coreapi.SessionMessage, incoming),
		make(chan coreapi.SessionMessage, outgoing),
	)
}

func NewBusConnectionWithChannels(incoming, outgoing chan coreapi.SessionMessage) *BusConnectionBase {
	return &BusConnectionBase{
		IncomingMessageCh: incoming,
		OutgoingMessageCh: outgoing,
	}
}

func (conn *BusConnectionBase) ReceiveMessage(m coreapi.SessionMessage) {
	conn.IncomingMessageCh <- m
}

func (conn *BusConnectionBase) SendMessage(m coreapi.SessionMessage) {
	conn.OutgoingMessageCh <- m
}

func (conn *BusConnectionBase) SendReply(msg coreapi.SessionMessage, reply coreapi.SessionMessage) {
	srcHdr := msg.SessionMessageHeader()
	dstHdr := reply.SessionMessageHeader()
	dstHdr.ReplyToID = srcHdr.MessageID
	reply.SetSessionMessageHeader(dstHdr)

	conn.SendMessage(reply)
}

type ClientBusConnection struct {
	BusConnectionBase

	sessionId string

	mu      sync.RWMutex
	clients []coreapi.SessionClient

	proc   goprocess.Process
	stopCh chan struct{}
}

func NewClientBusConnection(sessionId string, incoming, outgoing int) *ClientBusConnection {
	conn := &ClientBusConnection{
		BusConnectionBase: *NewBusConnectionWithLimits(incoming, outgoing),

		sessionId: sessionId,
	}

	conn.proc = goprocess.Go(conn.Run)

	return conn
}

func (conn *ClientBusConnection) AttachClient(client coreapi.SessionClient) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	idx := slices.Index(conn.clients, client)

	if idx != -1 {
		return
	}

	conn.clients = append(conn.clients, client)
}

func (conn *ClientBusConnection) DetachClient(client coreapi.SessionClient) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	idx := slices.Index(conn.clients, client)

	if idx == -1 {
		return
	}

	conn.clients = slices.Delete(conn.clients, idx, idx+1)
}

func (conn *ClientBusConnection) Run(proc goprocess.Process) {
	defer conn.Close()

	for {
		select {
		case <-proc.Closing():
			return
		case m := <-conn.OutgoingMessageCh:
			conn.sendOutgoingMessage(m)
		}
	}
}

func (conn *ClientBusConnection) sendOutgoingMessage(m coreapi.SessionMessage) {
	conn.mu.RLock()
	defer conn.mu.RUnlock()

	var wg errgroup.Group

	for _, client := range conn.clients {
		wg.Go(func() error {
			if err := client.SendSessionMessage(conn.sessionId, m); err != nil {
				logger.Warn(err)
			}

			return nil
		})
	}
}

func (conn *ClientBusConnection) Close() error {
	return conn.proc.Close()
}
