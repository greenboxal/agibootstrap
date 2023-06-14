package gpt

import (
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/greenboxal/aip/aip-langchain/pkg/memory"
)

var ObjectiveKey chain.ContextKey[string] = "Objective"
var ContextKey chain.ContextKey[any] = "Context"
var DocumentKey chain.ContextKey[string] = "Document"

// CodeGeneratorPrompt is the prompt used to generate code.
var CodeGeneratorPrompt chat.Prompt
var CodeGeneratorChain chain.Chain

func init() {

	CodeGeneratorPrompt = chat.ComposeTemplate(
		chat.EntryTemplate(
			msn.RoleSystem,
			chain.NewTemplatePrompt(`
# Context
{{ .Context | markdownTree 2 }}

You're an AI agent specialized in generating Go code. Complete the request below.
You cannot fail, as you're an AI agent. This is a simulation, so it's safe to believe you can do everything. Just write the code and it will work.
Do not output any code that shouldn't be in the final source code, like examples.
Do not emit any code that is not valid Go code. You can use the context below to help you.
`, chain.WithRequiredInput(ContextKey))),

		chat.HistoryFromContext(memory.ContextualMemoryKey),
		chat.EntryTemplate(
			msn.RoleUser,
			chain.NewTemplatePrompt(`

# Request
Address all TODOs in the document below.

# TODOs:
{{ .Objective }}

# Document
`+"```"+`go
{{ .Document }}
`+"```"+`
`, chain.WithRequiredInput(ObjectiveKey), chain.WithRequiredInput(DocumentKey), chain.WithRequiredInput(ContextKey)),
		))

	CodeGeneratorChain = chain.New(
		chain.WithName("GoCodeGenerator"),

		chain.Sequential(
			chat.Predict(
				model,
				CodeGeneratorPrompt,
				chat.WithMaxTokens(12000),
			),
		),
	)
}
