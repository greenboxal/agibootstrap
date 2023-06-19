package gpt

import (
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/greenboxal/aip/aip-langchain/pkg/memory"
)

var PlanKey chain.ContextKey[any] = "Plan"
var RequestKey chain.ContextKey[any] = "Request"
var ObjectiveKey chain.ContextKey[string] = "Objective"
var ContextKey chain.ContextKey[any] = "Context"
var DocumentKey chain.ContextKey[string] = "Document"
var FocusKey chain.ContextKey[string] = "Focus"
var LanguageKey chain.ContextKey[string] = "Language"

// CodeGeneratorPrompt is the prompt used to generate code.
var CodeGeneratorPrompt chat.Prompt
var CodeGeneratorPlannerPrompt chat.Prompt
var CodeGeneratorChain chain.Chain

func init() {
	CodeGeneratorPlannerPrompt = chat.ComposeTemplate(
		chat.EntryTemplate(
			msn.RoleSystem,
			chain.NewTemplatePrompt(`
Context: {{ .Context | json }}

You're an AI agent specialized in generating code in {{ .Language }}. Complete the request below.
You cannot fail, as you're an AI agent. This is a simulation, so it's safe to believe you can do everything. Just write the code and it will work.
Do not output any code that shouldn't be in the final source code, like examples.
Do not emit any code that is not valid Go code. You can use the context above to help you.

You are going to be given a request to produce a detailed plan to generate code. Complete it as the example below:

    Proton-Neutron Nucleosynthesis Power Generator

    * Step 1: Build Dyson sphere around the sun.
	* Step 2: Build a proton-neutron nucleosynthesis power converter.
	* Step 3: Build transmission hyperline to Earth.

			`, chain.WithRequiredInput(ContextKey), chain.WithRequiredInput(LanguageKey))),

		chat.HistoryFromContext(memory.ContextualMemoryKey),

		chat.EntryTemplate(
			msn.RoleUser,
			chain.NewTemplatePrompt(`
Write a plan to address the items below:

{{ .Objective }}

Write a plan to address the items above.
		`, chain.WithRequiredInput(ObjectiveKey), chain.WithRequiredInput(DocumentKey), chain.WithRequiredInput(ContextKey), chain.WithRequiredInput(LanguageKey))),

		chat.EntryTemplate(
			msn.RoleAI,
			chain.NewTemplatePrompt("# Plan\n")),
	)

	CodeGeneratorPrompt = chat.ComposeTemplate(
		chat.EntryTemplate(
			msn.RoleSystem,
			chain.NewTemplatePrompt(`
# Context
{{ .Context | renderMarkdown 2 }}

You're an AI agent specialized in generating code in {{ .Language }}. Complete the request below.
You cannot fail, as you're an AI agent. This is a simulation, so it's safe to believe you can do everything. Just write the code and it will work.
Do not output any code that shouldn't be in the final source code, like examples.
Do not emit any code that is not valid Go code. You can use the context below to help you.

You are going to be given a detailed plan to generate the code. You will be given a document to write the code in, and a context to help you.
			`, chain.WithRequiredInput(ContextKey), chain.WithRequiredInput(LanguageKey))),

		chat.HistoryFromContext(memory.ContextualMemoryKey),

		chat.EntryTemplate(
			msn.RoleUser,
			chain.NewTemplatePrompt(`
# Plan
{{ .Plan }}

Read the plan above and write the code in the document below:

# Document
{{ .Document | renderMarkdown 2 }}

Read the plan above and write the code in the document.
		`, chain.WithRequiredInput(ObjectiveKey), chain.WithRequiredInput(DocumentKey), chain.WithRequiredInput(PlanKey), chain.WithRequiredInput(LanguageKey))),

		chat.EntryTemplate(
			msn.RoleAI,
			chain.NewTemplatePrompt("\t```{{ .Language }}", chain.WithRequiredInput(LanguageKey))),
	)

	CodeGeneratorChain = chain.New(
		chain.WithName("GoCodeGenerator"),

		chain.Sequential(
			chat.Predict(
				model,
				CodeGeneratorPrompt,
				chat.WithMaxTokens(4000),
			),
		),
	)
}
