package client

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

type Client struct {
	sp inject.ServiceProvider
	vg *graphfs.VirtualGraph
}

func NewClient(sp inject.ServiceProvider, vg *graphfs.VirtualGraph) *Client {
	return &Client{sp: sp, vg: vg}
}

func (c *Client) BeginTransaction(ctx context.Context, options ...coreapi.TransactionOption) (coreapi.Transaction, error) {

}

func (c *Client) RunTransaction(ctx context.Context, fn coreapi.TransactionFunc, options ...coreapi.TransactionOption) (err error) {
	return coreapi.RunTransaction(ctx, c, fn, options...)
}

func (c *Client) Close() error {
	return c.vg.Close(context.Background())
}
