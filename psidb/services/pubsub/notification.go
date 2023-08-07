package pubsub

import "github.com/greenboxal/agibootstrap/pkg/psi"

type Notification struct {
	Ts   int64    `json:"ts"`
	Path psi.Path `json:"path"`
}
