package coreapi

import (
	"context"
	"fmt"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

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
