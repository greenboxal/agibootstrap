package agents

import (
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
)

type PromptToolSelection struct {
	Focus     *stdlib.Reference[psi.Node] `json:"focus"`
	Name      string                      `json:"name"`
	Arguments string                      `json:"arguments"`

	Tool PromptBuilderTool `json:"-"`
}

type PromptResponse struct {
	Raw     *gpt.Trace             `json:"raw"`
	Choices []PromptResponseChoice `json:"choices"`
}

type PromptResponseChoice struct {
	Index   int                         `json:"index"`
	Message *Message                    `json:"message"`
	Reason  openai.ChatCompletionReason `json:"reason"`

	Tool *PromptToolSelection `json:"tool"`
}
