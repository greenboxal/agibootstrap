package ws

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/jbenet/goprocess"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/pubsub"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Client struct {
	handler *Handler
	conn    *websocket.Conn
	proc    goprocess.Process

	midCounter atomic.Uint64

	outgoingCh chan []byte
	stopCh     chan struct{}

	subscriptions map[string]*pubsub.Subscription

	session coreapi.Session
}

func NewClient(
	handler *Handler,
	conn *websocket.Conn,
) *Client {
	return &Client{
		conn:          conn,
		handler:       handler,
		outgoingCh:    make(chan []byte, 256),
		stopCh:        make(chan struct{}),
		subscriptions: map[string]*pubsub.Subscription{},
	}
}

func (c *Client) SendMessage(msg *Message) error {
	msg.Mid = c.midCounter.Add(1)

	data, err := ipld.Encode(typesystem.Wrap(msg), dagjson.Encode)

	if err != nil {
		return err
	}

	c.outgoingCh <- data

	return nil
}

func (c *Client) SendSessionMessage(sessionId string, msg coreapi.SessionMessage) error {
	hdr := msg.SessionMessageHeader()
	hdr.SessionID = sessionId
	msg.SetSessionMessageHeader(hdr)

	return c.SendMessage(&Message{
		ReplyTo: hdr.ReplyToID,
		Session: &SessionMessage{
			SessionID: sessionId,
			Message:   msg,
		},
	})
}

func (c *Client) handleSession(msg Message) error {
	msg.Session.Message.SetSessionMessageHeader(coreapi.SessionMessageHeader{
		SessionID: msg.Session.SessionID,
		MessageID: msg.Mid,
		ReplyToID: msg.ReplyTo,
	})

	if m, ok := msg.Session.Message.(*coreapi.SessionMessageOpen); ok {
		if c.session != nil {
			if c.session.UUID() != msg.Session.SessionID {
				return c.SendMessage(&Message{
					ReplyTo: msg.Mid,
					Nack:    &NackMessage{},
				})
			}
		}

		if msg.Session.SessionID == "" && m.Config.SessionID != "" {
			msg.Session.SessionID = m.Config.SessionID
		} else if msg.Session.SessionID != "" && m.Config.SessionID == "" {
			m.Config.SessionID = msg.Session.SessionID
		}

		if (m.Config.KeepAliveTimeout == 0 && m.Config.Deadline.IsZero()) && !m.Config.Persistent {
			m.Config.KeepAliveTimeout = 30 * time.Second
		}

		if msg.Session.SessionID != m.Config.SessionID {
			return c.SendMessage(&Message{
				ReplyTo: msg.Mid,
				Nack:    &NackMessage{},
			})
		}

		if msg.Session.SessionID == "" {
			m.Config.SessionID = uuid.NewString()
			msg.Session.SessionID = m.Config.SessionID
		}

		sess := c.handler.sessionManager.GetOrCreateSession(m.Config)

		sess.AttachClient(c)

		c.session = sess
	}

	if c.session == nil {
		return c.SendMessage(&Message{
			ReplyTo: msg.Mid,
			Nack:    &NackMessage{},
		})
	}

	c.session.ReceiveMessage(msg.Session.Message)

	return nil
}

func (c *Client) handleSubscribe(msg Message) error {
	path, err := psi.ParsePath(msg.Subscribe.Topic)

	if err != nil {
		return err
	}

	if s := c.subscriptions[msg.Subscribe.Topic]; s != nil {
		if s.Pattern().Depth >= msg.Subscribe.Depth || msg.Subscribe.Depth == -1 {
			return nil
		} else {
			defer s.Close()
		}
	}

	s := c.handler.pubsub.Subscribe(pubsub.SubscriptionPattern{
		ID:    uuid.NewString(),
		Path:  path,
		Depth: msg.Subscribe.Depth,
	}, func(not pubsub.Notification) {
		if err := c.SendMessage(&Message{
			Notify: &NotificationMessage{
				Notification: not,
			},
		}); err != nil {
			c.handler.logger.Error(err)
		}
	})

	c.subscriptions[msg.Subscribe.Topic] = s

	return c.SendMessage(&Message{
		ReplyTo: msg.Mid,
		Ack:     &AckMessage{},
	})
}

func (c *Client) handleUnsubscribe(msg Message) error {
	s := c.subscriptions[msg.Subscribe.Topic]

	if s == nil {
		return nil
	}

	return c.SendMessage(&Message{
		ReplyTo: msg.Mid,
		Ack:     &AckMessage{},
	})
}

func (c *Client) handleMessage(message []byte) error {
	msgNode, err := ipld.DecodeUsingPrototype(message, dagjson.Decode, MessageType.IpldPrototype())

	if err != nil {
		return err
	}

	msg := typesystem.Unwrap(msgNode).(Message)

	if msg.Subscribe != nil {
		return c.handleSubscribe(msg)
	} else if msg.Unsubscribe != nil {
		return c.handleUnsubscribe(msg)
	} else if msg.Session != nil {
		return c.handleSession(msg)
	}

	if c.session != nil {
		c.session.KeepAlive()
	}

	return nil
}

func (c *Client) readPump(proc goprocess.Process) {
	defer func() {
		if c.outgoingCh != nil {
			close(c.outgoingCh)
			c.outgoingCh = nil
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		panic(err)
	}

	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		select {
		case _, _ = <-proc.Closing():
			return
		case _, _ = <-c.stopCh:
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.handler.logger.Error(err)
			}

			return
		}

		select {
		case _, _ = <-proc.Closing():
			return
		case _, _ = <-c.stopCh:
			return
		default:
		}

		if err := c.handleMessage(message); err != nil {
			c.handler.logger.Error(err)
		}
	}
}

func (c *Client) writePump(proc goprocess.Process) {
	defer func() {
		if c.stopCh != nil {
			close(c.stopCh)
			c.stopCh = nil
		}
	}()

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case _, _ = <-proc.Closing():
			return
		case _, _ = <-c.stopCh:
			return
		case message, ok := <-c.outgoingCh:
			if !ok {
				return
			}

			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.handler.logger.Error(err)
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)

			if err != nil {
				return
			}

			if _, err := w.Write(message); err != nil {
				return
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Close() error {
	if c.outgoingCh != nil {
		close(c.outgoingCh)
		c.outgoingCh = nil
	}

	if c.proc != nil {
		if err := c.proc.Close(); err != nil {
			return nil
		}
	}

	return nil
}

func (c *Client) teardown() error {
	if c.session != nil {
		c.session.DetachClient(c)
		c.session = nil
	}

	if err := c.conn.Close(); err != nil {
		return err
	}

	return nil
}

func (c *Client) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	c.proc = goprocess.Go(func(proc goprocess.Process) {
		proc.SetTeardown(c.teardown)

		proc.WaitFor(proc.Go(c.writePump))
		proc.WaitFor(proc.Go(c.readPump))

		select {
		case _, _ = <-proc.Closing():
			return
		case _, _ = <-c.stopCh:
			return
		}
	})
}
