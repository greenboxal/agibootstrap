package runtime

import (
	"github.com/greenboxal/agibootstrap/pkg/platform"
	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
)

type Platform struct {
	eventLoop       platform.EventLoop
	serviceProvider inject.ServiceProvider
}

func NewPlatform() platform.Platform {
	return &Platform{
		eventLoop:       NewEventLoop(),
		serviceProvider: inject.NewServiceProvider(),
	}
}

func (p *Platform) EventLoop() platform.EventLoop           { return p.eventLoop }
func (p *Platform) ServiceProvider() inject.ServiceProvider { return p.serviceProvider }

func Initialize() {
	platform.SetInstance(NewPlatform())
}
