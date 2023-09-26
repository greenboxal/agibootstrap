package repl

import (
	"context"
	"io"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/core/vm"
	"rogchap.com/v8go"
)

type StandardIO interface {
	Stdout() io.Writer
	Stderr() io.Writer
	Stdin() io.Reader
}

type StreamStandardIO struct {
	Out io.Writer
	Err io.Writer
	In  io.Reader
}

func (s StreamStandardIO) Stdout() io.Writer { return s.Out }
func (s StreamStandardIO) Stderr() io.Writer { return s.Err }
func (s StreamStandardIO) Stdin() io.Reader  { return s.In }

type ServerHandler struct {
	sess  coreapi.Session
	ctx   context.Context
	vmctx *vm.Context
	stdio StandardIO
}

func NewServerReplHandler(
	ctx context.Context,
	sess coreapi.Session,
	stdio StandardIO,
) *ServerHandler {
	sp := sess.ServiceProvider()
	iso := inject.Inject[*vm.Isolate](sess.ServiceProvider())
	vmctx := vm.NewContext(ctx, iso, sp)

	return &ServerHandler{
		sess:  sess,
		ctx:   ctx,
		vmctx: vmctx,
		stdio: stdio,
	}
}

func (c *ServerHandler) RunLine(ctx context.Context, line string) (any, error) {
	res, err := c.vmctx.Eval(line, "<repl>")

	if err != nil {
		return nil, err
	}

	json, err := v8go.JSONStringify(c.vmctx.Context(), res)

	if err != nil {
		return nil, err
	}

	return json, nil
}

func (c *ServerHandler) Close() error {
	return c.vmctx.Close()
}
