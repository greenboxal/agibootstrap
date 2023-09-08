package agents

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type PostStepHook func(ctx AgentContext, thought *thoughtdb.Thought) error

type Profile struct {
	psi.NodeBase

	Name        string
	Description any

	BaselineSystemPrompt string

	Rank     float64
	Priority float64

	Provides []string
	Requires []string

	PostStep PostStepHook
}

func BuildProfile(builder func(*Profile)) *Profile {
	p := &Profile{
		Rank:     1.0,
		Priority: 1.0,
	}

	p.Init(p)
	builder(p)
	return p
}

func (p *Profile) Clone() *Profile {
	clone := &Profile{
		Name:                 p.Name,
		Description:          p.Description,
		BaselineSystemPrompt: p.BaselineSystemPrompt,
		Rank:                 p.Rank,
		Priority:             p.Priority,
		Provides:             p.Provides,
		Requires:             p.Requires,
		PostStep:             p.PostStep,
	}

	clone.Init(clone)

	return clone
}
