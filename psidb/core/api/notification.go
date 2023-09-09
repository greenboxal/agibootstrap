package coreapi

import (
	"context"
	"fmt"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

func DispatchAction(
	ctx context.Context,
	from psi.Node,
	target psi.Node,
	iface psi.NodeInterface,
	action string,
	args any,
	options ...psi.NotificationOption,
) error {
	return Dispatch(ctx, psi.Notification{
		Notifier:  from.CanonicalPath(),
		Notified:  target.CanonicalPath(),
		Interface: iface.Name(),
		Action:    action,
		Argument:  args,
	}, options...)
}

func DispatchSelf(
	ctx context.Context,
	self psi.Node,
	iface psi.NodeInterface,
	action string,
	args any,
	options ...psi.NotificationOption,
) error {
	return Dispatch(ctx, psi.Notification{
		Notified:  self.CanonicalPath(),
		Notifier:  self.CanonicalPath(),
		Interface: iface.Name(),
		Action:    action,
		Argument:  args,
	}, options...)
}

func Dispatch(ctx context.Context, not psi.Notification, options ...psi.NotificationOption) error {
	tx := GetTransaction(ctx)

	if tx == nil {
		return fmt.Errorf("no transaction")
	}

	if len(options) > 0 {
		not = not.WithOptions(options...)
	}

	return tx.Notify(ctx, not)
}
