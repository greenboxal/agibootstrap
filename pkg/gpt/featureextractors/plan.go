package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type PlanStep struct {
	psi.NodeBase

	Name        string `json:"name" jsonschema:"title=Name,description=Name of the step."`
	Description string `json:"description" jsonschema:"title=Description,description=Description of the step."`
}

type PlanStepStatus struct {
	psi.NodeBase

	Step   *PlanStep       `json:"Step" jsonschema:"title=Step,description=Step that was completed."`
	Status *GoalCompletion `json:"Status" jsonschema:"title=Status,description=Status of the step."`
}

type Plan struct {
	psi.NodeBase

	Name  string     `json:"name" jsonschema:"title=Name,description=Name of the plan."`
	Steps []PlanStep `json:"steps" jsonschema:"title=Steps,description=List of steps in the plan."`
}

type PlanStatus struct {
	psi.NodeBase

	Plan       *Plan             `json:"Plan" jsonschema:"title=Plan,description=Plan that was completed."`
	StepStatus []*PlanStepStatus `json:"StepStatus" jsonschema:"title=StepStatus,description=Status of each step in the plan."`
	Status     *GoalCompletion   `json:"Status" jsonschema:"title=Status,description=Status of the plan."`
}

func QueryPlan(ctx context.Context, history []*thoughtdb.Thought) (Plan, error) {
	res, _, err := Reflect[Plan](ctx, ReflectOptions{
		History: history,

		Query: "The message above contains a plan. Please split the message with the relevant details for each individual step.",

		ExampleInput: `
This is a plan for X. Steps to accomplish X:

* Step 1: Do this.
* Step 2: Do that.
* Step 3: Do this and that.
* Step 4: Do this and that, and also this.
* Step 5: Do this and that, and also this and that.
`,

		ExampleOutput: Plan{
			Name: "X",
			Steps: []PlanStep{
				{
					Name:        "Step 1",
					Description: "Do this.",
				},
				{
					Name:        "Step 2",
					Description: "Do that.",
				},
				{
					Name:        "Step 3",
					Description: "Do this and that.",
				},
				{
					Name:        "Step 4",
					Description: "Do this and that, and also this.",
				},
				{
					Name:        "Step 5",
					Description: "Do this and that, and also this and that.",
				},
			},
		},
	})

	return res, err
}
