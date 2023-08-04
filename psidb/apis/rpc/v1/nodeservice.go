package rpcv1

import (
	"bytes"
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering/themes"
	"github.com/greenboxal/agibootstrap/psidb/core"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type NodeService struct {
	core *core.Core
}

func NewNodeService(core *core.Core) *NodeService {
	return &NodeService{core: core}
}

type GetNodeRequest struct {
	Path psi.Path `json:"path"`
}

type GetNodeResponse struct {
	Node psi.Node `json:"node,omitempty"`
}

func (ns *NodeService) GetNode(ctx context.Context, req *GetNodeRequest) (res *GetNodeResponse, err error) {
	err = ns.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) (err error) {
		n, err := tx.Resolve(ctx, req.Path)

		if err != nil {
			return err
		}

		res = &GetNodeResponse{Node: n}

		return nil
	})

	return
}

type RenderNodeRequest struct {
	Path   psi.Path `json:"path"`
	Format string   `json:"format"`
	View   string   `json:"view"`
}

type RenderNodeResponse struct {
	Rendered []byte `json:"rendered"`
}

func (ns *NodeService) RenderNode(ctx context.Context, req *RenderNodeRequest) (res *RenderNodeResponse, err error) {
	err = ns.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) (err error) {
		var buffer bytes.Buffer

		n, err := tx.Resolve(ctx, req.Path)

		if err != nil {
			return err
		}

		err = rendering.RenderNodeWithTheme(ctx, &buffer, themes.GlobalTheme, req.Format, req.View, n)

		if err != nil {
			return err
		}

		res.Rendered = buffer.Bytes()

		return nil
	})

	return
}
