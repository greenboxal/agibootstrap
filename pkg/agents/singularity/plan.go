package singularity

import (
	"github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt/featureextractors"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
)

type PlanStepResult struct {
	Step     *featureextractors.PlanStep
	Status   *featureextractors.PlanStepStatus
	Thoughts []*thoughtstream.Thought
}

type PlanStepHandler interface {
	HandlePlanStep(
		ctx agents.AgentContext,
		planStep *featureextractors.PlanStep,
		history ...*thoughtstream.Thought,
	) (*PlanStepResult, error)
}

type ResolvedPlanStep struct {
	*featureextractors.PlanStep

	PlanStepHandler
}

type ResolvedPlan struct {
	*featureextractors.Plan

	ResolvedSteps []*ResolvedPlanStep
}

type InstinctivePlanStepHandler struct{}

func (i *InstinctivePlanStepHandler) HandlePlanStep(
	ctx agents.AgentContext,
	planStep *featureextractors.PlanStep,
	history ...*thoughtstream.Thought,
) (*PlanStepResult, error) {
	ctx.WorldState().Set("current_plan_step", planStep)

	err := ctx.Agent().Step(ctx.Context())

	if err != nil {
		return nil, err
	}

	return &PlanStepResult{}, nil
}
