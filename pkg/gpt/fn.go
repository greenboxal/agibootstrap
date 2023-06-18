package gpt

import (
	"encoding/json"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
)

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
