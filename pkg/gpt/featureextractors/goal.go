package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
)

type GoalCompletion struct {
	Goal                   string `json:"goal" jsonschema:"title=Goal,description=Description of the goal."`
	Completed              bool   `json:"completed" jsonschema:"title=Completed,description=Whether the goal is completed."`
	CompletionStepsCurrent int    `json:"completionStepsCurrent" jsonschema:"title=Completion steps current,description=Number of steps completed."`
	CompletionStepsTotal   int    `json:"completionStepsTotal" jsonschema:"title=Completion steps total,description=Total number of steps."`
	Reasoning              string `json:"reasoning" jsonschema:"title=Reasoning,description=Reasoning for the completion."`
	Feedback               string `json:"feedback" jsonschema:"title=Feedback,description=Feedback for the completion."`
}

func QueryGoalCompletion(ctx context.Context, history []thoughtstream.Thought) (GoalCompletion, error) {
	res, _, err := Reflect[GoalCompletion](ctx, ReflectOptions{
		History: history,

		Query: "The message above contains a status report about a goal. Please tell me whether the goal was completed or not, and if ot, provide a progress/ETA, reasoning and feedback.",

		ExampleInput: "",
	})

	return res, err
}
