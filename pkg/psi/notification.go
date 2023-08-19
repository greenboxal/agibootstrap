package psi

import (
	"context"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"
)

type Confirmation struct {
	Xid   uint64 `json:"xid"`
	Rid   uint64 `json:"rid"`
	Nonce uint64 `json:"nonce"`
	Ok    bool   `json:"ok"`
}

type Notification struct {
	Nonce uint64 `json:"nonce"`

	Notifier  Path   `json:"notifier"`
	Notified  Path   `json:"notified"`
	Interface string `json:"interface"`
	Action    string `json:"action"`

	Argument any    `json:"-"`
	Params   []byte `json:"params,omitempty"`
}

func (n Notification) Apply(ctx context.Context, target Node) error {
	typ := target.PsiNodeType()
	iface := typ.Interface(n.Interface)

	if iface == nil {
		return errors.New("interface not found")
	}

	action := iface.Action(n.Action)

	if action == nil {
		return errors.New("action not found")
	}

	argsNode, err := ipld.DecodeUsingPrototype(n.Params, dagjson.Decode, action.RequestType().IpldPrototype())

	if err != nil {
		return err
	}

	_, err = action.Invoke(ctx, target, typesystem.Unwrap(argsNode))

	return err
}
