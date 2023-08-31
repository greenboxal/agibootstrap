package client

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jbenet/goprocess"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
)

var logger = logging.GetLogger("psidb/client")

type Client struct {
	mu sync.RWMutex

	transport Transport
	conn      Connection

	incomingCh chan Message
	tracker    *RpcResponseTracker

	midCounter atomic.Uint64
}

func NewClient(transport Transport) *Client {
	return &Client{
		transport: transport,
		tracker:   NewRpcResponseTracker(),
	}
}

func (c *Client) Connect(ctx context.Context) (Connection, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.connectUnlocked(ctx)
}

func (c *Client) getConnection(ctx context.Context) (Connection, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		c.mu.Lock()
		defer c.mu.Unlock()

		return c.connectUnlocked(ctx)
	}

	return conn, nil
}

func (c *Client) connectUnlocked(ctx context.Context) (Connection, error) {
	if c.conn != nil && c.conn.IsConnected() {
		return c.conn, nil
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return nil, err
		}

		c.conn = nil
	}

	c.incomingCh = make(chan Message, 16)

	conn, err := c.transport.Connect(ctx, c.incomingCh)

	if err != nil {
		return nil, err
	}

	c.conn = conn

	goprocess.Go(c.runMessagePump)

	return nil, err
}

func (c *Client) Disconnect() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}

		c.conn = nil
	}

	if c.incomingCh != nil {
		close(c.incomingCh)
		c.incomingCh = nil
	}

	return nil
}

func (c *Client) Close() error {
	return c.Disconnect()
}

func (c *Client) SendMessage(ctx context.Context, msg Message) error {
	conn, err := c.getConnection(ctx)

	if err != nil {
		return err
	}

	return conn.SendMessage(ctx, msg)
}

func (c *Client) runMessagePump(proc goprocess.Process) {
	for {
		select {
		case <-proc.Closing():
			return

		case msg, ok := <-c.incomingCh:
			if !ok {
				return
			}

			if err := c.handleIncomingMessage(msg); err != nil {
				logger.Error(err)
			}
		}
	}
}

func (c *Client) handleIncomingMessage(msg Message) error {
	if msg.GetMessageHeader().ReplyToID != 0 {
		c.tracker.AcceptMessage(msg)
	}

	return nil
}

func (c *Client) getNextMessageID() uint64 {
	return c.midCounter.Add(1)
}
