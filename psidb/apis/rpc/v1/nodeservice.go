package rpcv1

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"

	"github.com/greenboxal/agibootstrap/psidb/typesystem"

	"github.com/greenboxal/agibootstrap/psidb/core"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering/themes"
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

type CallNodeActionRequest struct {
	Path      psi.Path        `json:"path"`
	Interface string          `json:"interface"`
	Action    string          `json:"action"`
	Args      json.RawMessage `json:"args"`
}

type CallNodeActionResponse struct {
	Node   psi.Node `json:"node,omitempty"`
	Result any      `json:"result,omitempty"`
}

func (ns *NodeService) CallNodeAction(ctx context.Context, req *CallNodeActionRequest) (res *CallNodeActionResponse, err error) {
	res = &CallNodeActionResponse{}

	err = ns.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) (err error) {
		var actionRequest any

		n, err := tx.Resolve(ctx, req.Path)

		if err != nil {
			return err
		}

		res.Node = n

		typ := n.PsiNodeType()
		iface := typ.Interface(req.Interface)

		if iface == nil {
			return errors.New("interface not found")
		}

		action := iface.Action(req.Action)

		if action == nil {
			return errors.New("action not found")
		}

		if action.RequestType() != nil {
			argsNode, err := ipld.DecodeUsingPrototype(req.Args, dagjson.Decode, action.RequestType().IpldPrototype())

			if err != nil {
				return err
			}

			actionRequest = typesystem.Unwrap(argsNode)
		}

		ctx, span := tracer.Start(ctx, "NodeService.CallNodeAction")
		span.SetAttributes(semconv.ServiceName("NodeRunner"))
		span.SetAttributes(semconv.RPCSystemKey.String("psidb-node"))
		span.SetAttributes(semconv.RPCService(iface.Name()))
		span.SetAttributes(semconv.RPCMethod(action.Name()))
		defer span.End()

		result, err := action.Invoke(ctx, n, actionRequest)

		if err != nil {
			return err
		}

		res.Result = result

		return nil
	})

	return
}
