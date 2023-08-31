package vm

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"rogchap.com/v8go"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/typesystem"
)

type VirtualNode struct {
	psi.NodeBase

	UUID string           `json:"uuid"`
	Data *json.RawMessage `json:"data"`

	ctx   *Context
	value *v8go.Value
}

var VirtualNodeType = typesystem.TypeOf((*VirtualNode)(nil))

func NewVirtualNode(ctx *Context, uuid string) *VirtualNode {
	vn := &VirtualNode{
		UUID: uuid,
		ctx:  ctx,
	}

	vn.Init(vn)

	return vn
}

func (on *VirtualNode) PsiNodeName() string {
	return on.UUID
}

func (on *VirtualNode) Init(self psi.Node, options ...psi.NodeInitOption) {
	if on.UUID == "" {
		on.UUID = uuid.NewString()
	}

	on.NodeBase.Init(self, options...)

	if err := on.updateValue(); err != nil {
		panic(err)
	}
}

func (on *VirtualNode) OnUpdate(ctx context.Context) error {
	return on.updateValue()
}

func (on *VirtualNode) updateValue() error {
	if on.ctx != nil {
		if on.value == nil {
			var data []byte

			if on.Data != nil {
				data = *on.Data
			}

			if len(data) == 0 {
				tmpl := v8go.NewObjectTemplate(on.ctx.iso.iso)
				obj, err := tmpl.NewInstance(on.ctx.ctx)

				if err != nil {
					return nil
				}

				on.value = obj.Value
			} else {
				value, err := v8go.JSONParse(on.ctx.ctx, string(data))

				if err != nil {
					return nil
				}

				on.value = value
			}
		} else if on.value != nil {
			data, err := v8go.JSONStringify(on.ctx.ctx, on.value)

			if err != nil {
				return nil
			}

			rm := json.RawMessage(data)
			on.Data = &rm
		}
	}

	return nil
}
