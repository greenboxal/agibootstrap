package planner

import (
	agents2 "github.com/greenboxal/agibootstrap/pkg/gpt/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt/promptml"
)

func BuildRootPrompt() agents2.AgentPromptFunc {
	return agents2.Tml(func(ctx agents2.AgentContext) promptml.Parent {

	})
}
