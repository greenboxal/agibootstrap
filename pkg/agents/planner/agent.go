package planner

import (
	"github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt/featureextractors"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
)

type Agent struct {
	agents.AgentBase
}

type WorldState interface {
	agents.WorldState

	GetPlan() *featureextractors.Plan
	UpdatePlan(plan *featureextractors.Plan)
}

func NewAgent(
	profile *agents.Profile,
	repo thoughtstream.Resolver,
	log *thoughtstream.ThoughtLog,
	worldState agents.WorldState,
) *Agent {
	a := &Agent{}

	a.Init(a, profile, repo, log, worldState)

	return a
}
