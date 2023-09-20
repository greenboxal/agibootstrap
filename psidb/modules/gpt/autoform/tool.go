package autoform

import (
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

type PromptBuilderTool interface {
	ToolName() string
	ToolDefinition() *openai.FunctionDefinition
}

type simpleTool struct {
	definition *openai.FunctionDefinition
}

func WrapTool(def *openai.FunctionDefinition) PromptBuilderTool {
	return simpleTool{definition: def}
}

func (s simpleTool) ToolName() string                           { return s.definition.Name }
func (s simpleTool) ToolDefinition() *openai.FunctionDefinition { return s.definition }

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
	Message *chat.Message               `json:"message"`
	Reason  openai.ChatCompletionReason `json:"reason"`

	Tool *PromptToolSelection `json:"tool"`
}
