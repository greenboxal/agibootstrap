package singularity

import (
	"github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt/featureextractors"
)

type Scheduler interface {
	NextSpeaker(ctx agents.AgentContext, candidates ...agents.Agent) (agents.Agent, error)
}

type RoundRobinScheduler struct {
	current int
}

func (r *RoundRobinScheduler) NextSpeaker(ctx agents.AgentContext, candidates ...agents.Agent) (agents.Agent, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	if r.current >= len(candidates) {
		r.current = 0
	}

	next := candidates[r.current]
	r.current++

	return next, nil
}

type DefaultModeScheduler struct{}

func (s *DefaultModeScheduler) NextSpeaker(ctx agents.AgentContext, candidates ...agents.Agent) (agents.Agent, error) {
	plan := agents.GetState(ctx.WorldState(), CtxPlannerPlan)
	nextSpeaker, err := featureextractors.PredictNextSpeaker(ctx.Context(), plan, ctx.Log().Messages()...)

	if err != nil {
		return nil, err
	}

	for _, candidate := range candidates {
		if candidate.Profile().Name == nextSpeaker.NextSpeaker {
			return candidate, nil
		}
	}

	return ctx.Agent(), nil
}

type TaskPositiveScheduler struct{}

func (s *TaskPositiveScheduler) NextSpeaker(ctx agents.AgentContext, candidates ...agents.Agent) (agents.Agent, error) {
	plan := agents.GetState(ctx.WorldState(), CtxPlannerPlan)
	nextSpeaker, err := featureextractors.PredictNextSpeaker(ctx.Context(), plan, ctx.Log().Messages()...)

	if err != nil {
		return nil, err
	}

	for _, candidate := range candidates {
		if candidate.Profile().Name == nextSpeaker.NextSpeaker {
			return candidate, nil
		}
	}

	return ctx.Agent(), nil
}
