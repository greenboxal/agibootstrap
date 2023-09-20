package jukebox

import (
	`github.com/jbenet/goprocess`
	`gitlab.com/gomidi/midi/v2/drivers`
)

type CommandPacket struct {
	Context CommandContext
	Command EvaluableCommand
}

type Connection struct {
	channel uint8

	out drivers.Out

	proc goprocess.Process
	ch   chan CommandPacket

	ctx CommandContext
}

func NewConnection(channel uint8, out drivers.Out) *Connection {
	conn := &Connection{
		channel: channel,
		out:     out,

		ch: make(chan CommandPacket, 256),
	}

	conn.proc = goprocess.Go(conn.run)

	return conn
}

func (c *Connection) run(proc goprocess.Process) {
	proc.SetTeardown(c.teardown)

	for {
		select {
		case <-proc.Closing():
			return
		case pkt := <-c.ch:
			if err := c.Evaluate(pkt.Context, pkt.Command); err != nil {
				logger.Error(err)
			}
		}
	}
}

func (c *Connection) Evaluate(ctx CommandContext, cmd EvaluableCommand) error {
	c.ctx.Channel = c.channel
	c.ctx.Out = c.out

	return cmd.Evaluate(&c.ctx)
}

func (c *Connection) Close() error {
	return c.proc.Close()
}

func (c *Connection) teardown() error {
	return c.out.Close()
}

func (c *Connection) Enqueue(ctx CommandContext, cmd EvaluableCommand) {
	c.ch <- CommandPacket{
		Context: ctx,
		Command: cmd,
	}
}
