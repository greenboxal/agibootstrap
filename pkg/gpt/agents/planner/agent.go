package planner

import (
	agents2 "github.com/greenboxal/agibootstrap/pkg/gpt/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt/featureextractors"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
)

type Agent struct {
	agents2.AgentBase
}

type WorldState interface {
	agents2.WorldState

	GetPlan() *featureextractors.Plan
	UpdatePlan(plan *featureextractors.Plan)
}

func NewAgent(
	profile *agents2.Profile,
	repo thoughtdb.Resolver,
	log *thoughtdb.ThoughtLog,
	worldState agents2.WorldState,
) *Agent {
	a := &Agent{}

	a.Init(a, profile, repo, log, worldState)

	return a
}
