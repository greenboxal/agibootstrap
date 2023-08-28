package psi

import (
	"context"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"

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

	SessionID string             `json:"session_id"`
	TraceID   *NotificationTrace `json:"trace_id"`

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

type NotificationTrace struct {
	TraceID    string `json:"T,omitempty`
	SpanID     string `json:"S,omitempty"`
	TraceFlags int    `json:"F,omitempty"`
	TraceState string `json:"ST,omitempty"`
	Remote     bool   `json:"R,omitempty"`
}

func (n Notification) Apply(ctx context.Context, target Node) (any, error) {
	var arg any

	if n.TraceID != nil {
		cfg := trace.SpanContextConfig{}
		cfg.TraceID, _ = trace.TraceIDFromHex(n.TraceID.TraceID)
		cfg.SpanID, _ = trace.SpanIDFromHex(n.TraceID.SpanID)
		cfg.TraceState, _ = trace.ParseTraceState(n.TraceID.TraceState)
		cfg.TraceFlags = trace.TraceFlags(n.TraceID.TraceFlags)
		cfg.Remote = n.TraceID.Remote

		spanCtx := trace.NewSpanContext(cfg)
		ctx = trace.ContextWithRemoteSpanContext(ctx, spanCtx)
	}

	typ := target.PsiNodeType()
	iface := typ.Interface(n.Interface)

	if iface == nil {
		return nil, errors.New("interface not found")
	}

	action := iface.Action(n.Action)

	if action == nil {
		return nil, errors.New("action not found")
	}

	if action.RequestType() != nil {
		argsNode, err := ipld.DecodeUsingPrototype(n.Params, dagjson.Decode, action.RequestType().IpldPrototype())

		if err != nil {
			return nil, err
		}

		arg = typesystem.Unwrap(argsNode)
	}

	return action.Invoke(ctx, target, arg)
}
