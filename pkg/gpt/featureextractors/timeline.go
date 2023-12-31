package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type TimelineStep struct {
	psi.NodeBase

	Title string `json:"title" jsonschema:"title=Title,description=Title of the timeline step."`
	Body  string `json:"body" jsonschema:"title=Body,description=Body of the timeline step."`
}

type Timeline struct {
	psi.NodeBase

	Title string         `json:"title" jsonschema:"title=Title,description=Title of the timeline."`
	Steps []TimelineStep `json:"steps" jsonschema:"title=Steps,description=Steps of the timeline."`
}

func QueryTimeline(ctx context.Context, history ...*thoughtstream.Thought) (Timeline, error) {
	res, _, err := Reflect[Timeline](ctx, ReflectOptions{
		History: history,

		Query: "Summarize the chat history above and create a timeline about the events happening.",
	})

	return res, err
}
