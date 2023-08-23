package psi

import (
	"context"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"
)

type PromiseHandle struct {
	Xid   uint64 `json:"xid"`
	Nonce uint64 `json:"nonce"`
}

func (ph PromiseHandle) Wait(count int) Promise {
	return Promise{
		PromiseHandle: ph,
		Count:         count,
	}
}

func (ph PromiseHandle) Signal(count int) Promise {
	return Promise{
		PromiseHandle: ph,
		Count:         -count,
	}
}

type Promise struct {
	PromiseHandle

	Count int `json:"c"`
}

type Confirmation struct {
	Xid   uint64 `json:"xid"`
	Rid   uint64 `json:"rid"`
	Nonce uint64 `json:"nonce"`
	Ok    bool   `json:"ok"`
}

type Notification struct {
	Nonce uint64 `json:"nonce"`

	Notifier Path `json:"notifier"`
	Notified Path `json:"notified"`

	Interface string `json:"interface"`
	Action    string `json:"action"`

	Argument any    `json:"-"`
	Params   []byte `json:"params,omitempty"`

	Observers    []Promise `json:"observers,omitempty"`
	Dependencies []Promise `json:"dependencies,omitempty"`
}

type NotificationOption func(*Notification)

func WithNotifier(path Path) NotificationOption {
	return func(n *Notification) {
		n.Notifier = path
	}
}

func WithNotified(path Path) NotificationOption {
	return func(n *Notification) {
		n.Notified = path
	}
}

func WithObservers(observers ...Promise) NotificationOption {
	return func(n *Notification) {
		n.Observers = append(n.Observers, observers...)
	}
}

func WithDependencies(dependencies ...Promise) NotificationOption {
	return func(n *Notification) {
		n.Dependencies = append(n.Dependencies, dependencies...)
	}
}

func (n Notification) WithOptions(options ...NotificationOption) Notification {
	for _, option := range options {
		option(&n)
	}

	return n
}

func (n Notification) Apply(ctx context.Context, target Node) error {
	var arg any

	typ := target.PsiNodeType()
	iface := typ.Interface(n.Interface)

	if iface == nil {
		return errors.New("interface not found")
	}

	action := iface.Action(n.Action)

	if action == nil {
		return errors.New("action not found")
	}

	if action.RequestType() != nil {
		argsNode, err := ipld.DecodeUsingPrototype(n.Params, dagjson.Decode, action.RequestType().IpldPrototype())

		if err != nil {
			return err
		}

		arg = typesystem.Unwrap(argsNode)
	}

	_, err := action.Invoke(ctx, target, arg)

	return err
}
