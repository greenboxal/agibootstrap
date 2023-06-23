package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/agents"
)

type Routing struct {
	Recipients map[string][]string `json:"recipients" jsonschema:"title=Parties,description=Key: Name of the user/agent/subject/recipient. Value: List of relevant details of the message for the given key."`
}

func QueryMessageRouting(ctx context.Context, history []agents.Message) (Routing, error) {
	res, _, err := Reflect[Routing](ctx, ReflectOptions{
		History: history,

		Query: "The message above contains requests, commands, questions or replies to one or more users/agents/recipients. Please split the message with the relevant details for each individual subject.",

		ExampleInput: "Everyone pay attention! Bob do this and do that, and have Carol assist you. Carol also learn Y meanwhile. Alice write some reports afterwards.",

		ExampleOutput: Routing{
			Recipients: map[string][]string{
				"Bob":   {"Do this", "Do that"},
				"Carol": {"Assist Bob on X", "Learn Y"},
				"Alice": {"Write some reports"},
				"*":     {"Everyone please pay attention!"},
			},
		},
	})

	return res, err
}
