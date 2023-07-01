package planner

import (
	"github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt/promptml"
)

func BuildRootPrompt() agents.AgentPromptFunc {
	return agents.Tml(func(ctx agents.AgentContext) promptml.Parent {

	})
}
