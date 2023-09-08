package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Objective struct {
	psi.NodeBase

	Name        string   `json:"Name" jsonschema:"title=Name,description=Name of the objective."`
	Description string   `json:"Description" jsonschema:"title=Description,description=Description of the objective."`
	Keywords    []string `json:"Keywords" jsonschema:"title=Keywords,description=Keywords of the objective."`

	PositivePlan     Plan `json:"PositivePlan" jsonschema:"title=Positive plan,description=Plan to achieve the objective."`
	NegativePlan     Plan `json:"NegativePlan" jsonschema:"title=Negative plan,description=Plan to avoid the objective."`
	VerificationPlan Plan `json:"VerificationPlan" jsonschema:"title=Verification plan,description=Plan to verify if the objective was completed successfully."`

	Status GoalCompletion `json:"Status" jsonschema:"title=Status,description=Status of the goal."`
}

func QueryObjective(ctx context.Context, history []*thoughtdb.Thought) (Objective, error) {
	res, _, err := Reflect[Objective](ctx, ReflectOptions{
		History: history,

		Query: "Greate a new objective. Please provide a name, description, keywords, positive plan, negative plan and verification plan.",

		ExampleInput: "",
	})

	return res, err
}

type GoalCompletion struct {
	psi.NodeBase

	Goal                   string `json:"goal" jsonschema:"title=Goal,description=Description of the goal."`
	Completed              bool   `json:"completed" jsonschema:"title=Completed,description=Whether the goal is completed."`
	CompletionStepsCurrent int    `json:"completionStepsCurrent" jsonschema:"title=Completion steps current,description=Number of steps completed."`
	CompletionStepsTotal   int    `json:"completionStepsTotal" jsonschema:"title=Completion steps total,description=Total number of steps."`
	Reasoning              string `json:"reasoning" jsonschema:"title=Reasoning,description=Reasoning for the completion."`
	Feedback               string `json:"feedback" jsonschema:"title=Feedback,description=Feedback for the completion."`
}

func QueryGoalCompletion(ctx context.Context, history []*thoughtdb.Thought) (GoalCompletion, error) {
	res, _, err := Reflect[GoalCompletion](ctx, ReflectOptions{
		History: history,

		Query: "The message above contains a status report about a goal. Please tell me whether the goal was completed or not, and if ot, provide a progress/ETA, reasoning and feedback.",

		ExampleInput: "",
	})

	return res, err
}
