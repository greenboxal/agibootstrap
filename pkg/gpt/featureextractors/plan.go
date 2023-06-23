package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/agents"
)

type PlanStep struct {
	Name        string `json:"name" jsonschema:"title=Name,description=Name of the step."`
	Description string `json:"description" jsonschema:"title=Description,description=Description of the step."`
}

type Plan struct {
	Name  string     `json:"name" jsonschema:"title=Name,description=Name of the plan."`
	Steps []PlanStep `json:"steps" jsonschema:"title=Steps,description=List of steps in the plan."`
}

func QueryPlan(ctx context.Context, history []agents.Message) (Plan, error) {
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
