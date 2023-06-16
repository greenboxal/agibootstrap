package pair

import (
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
)

type Agent struct {
	nudgeChain    chain.Chain
	critiqueChain chain.Chain
	preemptChain  chain.Chain
}

// NewAgent returns a new Agent based on the provided language model.
func NewAgent(model chat.LanguageModel) *Agent {
	return &Agent{
		nudgeChain:    NewNudgeChain(model),
		critiqueChain: NewCritiqueChain(model),
		preemptChain:  NewPreemptChain(model),
	}
}

func (a *Agent) Run() {
	// TODO: Implement this method
	// Implement the logic for running the agent here
}
