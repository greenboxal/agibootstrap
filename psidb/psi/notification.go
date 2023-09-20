package psi

import (
	"context"
	"encoding/json"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	propagation2 "go.opentelemetry.io/otel/propagation"

	"github.com/greenboxal/agibootstrap/psidb/typesystem"
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

	SessionID string   `json:"session_id"`
	TraceID   string   `json:"trace_id"`
	TraceTags []string `json:"trace_tags"`

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

var ctxKeyTraceTags = &struct{ traceTags string }{traceTags: "trace_tags"}

func GetTraceTags(ctx context.Context) []string {
	v := ctx.Value(ctxKeyTraceTags)

	if v == nil {
		return nil
	}

	return v.([]string)
}

func AppendTraceTags(ctx context.Context, tags ...string) context.Context {
	current := GetTraceTags(ctx)

	return context.WithValue(ctx, ctxKeyTraceTags, append(current, tags...))
}

func (n Notification) Apply(ctx context.Context, target Node) (any, error) {
	var arg any

	if n.TraceID != "" {
		carrier := propagation2.MapCarrier{}

		if err := json.Unmarshal([]byte(n.TraceID), &carrier); err != nil {
			return nil, err
		}

		propagator := otel.GetTextMapPropagator()
		ctx = propagator.Extract(ctx, carrier)
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
