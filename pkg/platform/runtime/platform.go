package runtime

import "github.com/greenboxal/agibootstrap/pkg/platform"

type Platform struct {
	eventLoop platform.EventLoop
}

func NewPlatform() platform.Platform {
	return &Platform{}
}

func (p *Platform) EventLoop() platform.EventLoop { return p.eventLoop }

func Initialize() {
	platform.SetInstance(NewPlatform())
}
