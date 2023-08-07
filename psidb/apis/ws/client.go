package ws

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"

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

	midCounter atomic.Uint64
	outgoingCh chan []byte

	subscriptions map[string]*pubsub.Subscription
}

func NewClient(
	handler *Handler,
	conn *websocket.Conn,
) *Client {
	return &Client{
		conn:          conn,
		handler:       handler,
		outgoingCh:    make(chan []byte, 256),
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

func (c *Client) handleSubscribe(ctx context.Context, msg Message) error {
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

func (c *Client) handleUnsubscribe(ctx context.Context, msg Message) error {
	s := c.subscriptions[msg.Subscribe.Topic]

	if s == nil {
		return nil
	}

	return c.SendMessage(&Message{
		ReplyTo: msg.Mid,
		Ack:     &AckMessage{},
	})
}

func (c *Client) handleMessage(ctx context.Context, message []byte) error {
	msgNode, err := ipld.DecodeUsingPrototype(message, dagjson.Decode, MessageType.IpldPrototype())

	if err != nil {
		return err
	}

	msg := typesystem.Unwrap(msgNode).(Message)

	if msg.Subscribe != nil {
		return c.handleSubscribe(ctx, msg)
	} else if msg.Unsubscribe != nil {
		return c.handleUnsubscribe(ctx, msg)
	}

	return nil
}

func (c *Client) readPump(proc goprocess.Process) {
	ctx := goprocessctx.OnClosingContext(proc)

	c.conn.SetReadLimit(maxMessageSize)

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		panic(err)
	}

	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.handler.logger.Error(err)
			}

			break
		}

		if err := c.handleMessage(ctx, message); err != nil {
			c.handler.logger.Error(err)

			break
		}
	}
}

func (c *Client) writePump(proc goprocess.Process) {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()

		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.outgoingCh:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.handler.logger.Error(err)
				return
			}

			if !ok {
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)

			if err != nil {
				c.handler.logger.Error(err)
				break
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
	close(c.outgoingCh)

	return c.conn.Close()
}
