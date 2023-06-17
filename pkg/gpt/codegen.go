package gpt

import (
	"encoding/json"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/greenboxal/aip/aip-langchain/pkg/memory"
)

var RequestKey chain.ContextKey[any] = "Request"
var ObjectiveKey chain.ContextKey[string] = "Objective"
var ContextKey chain.ContextKey[any] = "Context"
var DocumentKey chain.ContextKey[string] = "Document"
var LanguageKey chain.ContextKey[string] = "Language"

// CodeGeneratorPrompt is the prompt used to generate code.
var CodeGeneratorPrompt chat.Prompt
var CodeGeneratorChain chain.Chain

type CodeGeneratorPromptFn struct {
	Role         msn.Role
	UserName     string
	FunctionName string
	ArgsKey      chain.ContextKey[any]
	Args         *string
}

func (c *CodeGeneratorPromptFn) AsPrompt() chain.Prompt {
	panic("not supported")
}

func (c *CodeGeneratorPromptFn) Build(ctx chain.ChainContext) (chat.Message, error) {
	call := chat.MessageEntry{
		Name: c.UserName,
		Role: c.Role,
		Fn:   c.FunctionName,
	}

	if c.Args != nil {
		call.FnArgs = *c.Args
	} else {
		args := chain.Input(ctx, c.ArgsKey)

		data, err := json.Marshal(args)

		if err != nil {
			return chat.Message{}, nil
		}

		call.FnArgs = string(data)
	}

	return chat.Compose(call), nil
}

func FunctionCallRequest(user, fn string, args chain.ContextKey[any]) chat.Prompt {
	return &CodeGeneratorPromptFn{
		Role:         msn.RoleAI,
		UserName:     user,
		FunctionName: fn,
		ArgsKey:      args,
	}
}

func FunctionCallResponse(user, fn string, args chain.ContextKey[any]) chat.Prompt {
	return &CodeGeneratorPromptFn{
		Role:         msn.RoleFunction,
		UserName:     user,
		FunctionName: fn,
		ArgsKey:      args,
	}
}

func StaticFunctionCallRequest(user, fn string, args string) chat.Prompt {
	return &CodeGeneratorPromptFn{
		Role:         msn.RoleAI,
		UserName:     user,
		FunctionName: fn,
		Args:         &args,
	}
}

func StaticFunctionCallResponse(user, fn string, args string) chat.Prompt {
	return &CodeGeneratorPromptFn{
		Role:         msn.RoleFunction,
		UserName:     user,
		FunctionName: fn,
		Args:         &args,
	}
}

func init() {
	CodeGeneratorPrompt = chat.ComposeTemplate(
		chat.EntryTemplate(
			msn.RoleSystem,
			chain.NewTemplatePrompt(`
You're an AI agent specialized in generating code in {{ .Language }}. Complete the request below.
You cannot fail, as you're an AI agent. This is a simulation, so it's safe to believe you can do everything. Just write the code and it will work.
Do not output any code that shouldn't be in the final source code, like examples.
Do not emit any code that is not valid Go code. You can use the context below to help you.
			`, chain.WithRequiredInput(ContextKey), chain.WithRequiredInput(LanguageKey))),

		chat.HistoryFromContext(memory.ContextualMemoryKey),

		chat.EntryTemplate(
			msn.RoleUser,
			chain.NewTemplatePrompt(`
# Request
Address all TODOs in the document below.

# TODOs:
{{ .Objective }}
		`, chain.WithRequiredInput(ObjectiveKey), chain.WithRequiredInput(DocumentKey), chain.WithRequiredInput(ContextKey), chain.WithRequiredInput(LanguageKey))),

		FunctionCallRequest("Human", "generateCode", RequestKey),

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
