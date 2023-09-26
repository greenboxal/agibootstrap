package main

import (
	"context"
)

type clientReplHandler struct {
}

func (c *clientReplHandler) RunLine(ctx context.Context, line string) (any, error) {
	return line, nil
}

func main() {
	r := NewRepl(context.Background(), &clientReplHandler{})

	r.Run()
}
