package main

import (
	"context"
	"os"
	"path"

	"github.com/go-errors/errors"
	"github.com/reeflective/readline"
)

type ReplHandler interface {
	RunLine(ctx context.Context, line string) (any, error)
}

type Repl struct {
	ctx    context.Context
	cancel context.CancelFunc

	sh      *readline.Shell
	handler ReplHandler

	closed bool
}

func NewRepl(ctx context.Context, handler ReplHandler) *Repl {
	r := &Repl{}

	r.handler = handler
	r.ctx, r.cancel = context.WithCancel(ctx)
	r.sh = readline.NewShell()

	if home := os.Getenv("HOME"); home != "" {
		histPath := path.Join(home, ".psic_history")

		if _, err := os.Stat(histPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if fd, err := os.Create(histPath); err != nil {
					panic(err)
				} else {
					if err := fd.Close(); err != nil {
						panic(err)
					}
				}
			} else {
				panic(err)
			}
		}
		hist, err := readline.NewHistoryFromFile(path.Join(home, ".psic_history"))

		if err != nil {
			panic(err)
		}

		r.sh.History.Add("user", hist)
	}

	r.sh.Prompt.Primary(func() string {
		return ">>> "
	})

	return r
}

func (r *Repl) Run() {
	for !r.closed {
		line, err := r.sh.Readline()

		if err != nil {
			panic(err)
		}

		if err = r.runLine(r.ctx, line); err != nil {
			r.handleError(err)
		}
	}
}

func (r *Repl) runLine(ctx context.Context, line string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.Wrap(e, 1)
		}
	}()

	result, err := r.handler.RunLine(ctx, line)

	if err != nil {
		return err
	}

	_, err = r.sh.Printf("%s", result)

	return err
}

func (r *Repl) handleError(err error) {
	_, e := r.sh.Printf("error: %s", err)

	if e != nil {
		panic(e)
	}
}

func (r *Repl) Close() error {
	r.closed = true
	r.cancel()

	return nil
}
