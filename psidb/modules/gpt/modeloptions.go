package gpt

import "github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

type ModelOptions struct {
	Model            *string        `json:"model" jsonschema:"title=Model,description=The model used for the conversation"`
	MaxTokens        *int           `json:"max_tokens" jsonschema:"title=Max Tokens,description=The maximum number of tokens for the model"`
	Temperature      *float32       `json:"temperature" jsonschema:"title=Temperature,description=The temperature setting for the model"`
	TopP             *float32       `json:"top_p" jsonschema:"title=Top P,description=The top P setting for the model"`
	FrequencyPenalty *float32       `json:"frequency_penalty" jsonschema:"title=Frequency Penalty,description=The frequency penalty setting for the model"`
	PresencePenalty  *float32       `json:"presence_penalty" jsonschema:"title=Presence Penalty,description=The presence penalty setting for the model"`
	Stop             []string       `json:"stop" jsonschema:"title=Stop,description=The stop words for the model"`
	LogitBias        map[string]int `json:"logit_bias" jsonschema:"title=Logit Bias,description=The logit bias setting for the model"`

	ForceFunctionCall *string `json:"force_function_call"`
}

func (o ModelOptions) MergeWith(opts ModelOptions) ModelOptions {
	if opts.Model != nil {
		o.Model = opts.Model
	}

	if opts.MaxTokens != nil {
		o.MaxTokens = opts.MaxTokens
	}

	if opts.Temperature != nil {
		o.Temperature = opts.Temperature
	}

	if opts.TopP != nil {
		o.TopP = opts.TopP
	}

	if opts.FrequencyPenalty != nil {
		o.FrequencyPenalty = opts.FrequencyPenalty
	}

	if opts.PresencePenalty != nil {
		o.PresencePenalty = opts.PresencePenalty
	}

	if opts.Stop != nil {
		o.Stop = opts.Stop
	}

	if opts.LogitBias != nil {
		o.LogitBias = opts.LogitBias
	}

	if opts.ForceFunctionCall != nil {
		o.ForceFunctionCall = opts.ForceFunctionCall
	}

	return o
}

func (o ModelOptions) Apply(req *openai.ChatCompletionRequest) {
	if o.Model != nil {
		req.Model = *o.Model
	}

	if o.MaxTokens != nil {
		req.MaxTokens = *o.MaxTokens
	}

	if o.Temperature != nil {
		req.Temperature = *o.Temperature
	}

	if o.TopP != nil {
		req.TopP = *o.TopP
	}

	if o.FrequencyPenalty != nil {
		req.FrequencyPenalty = *o.FrequencyPenalty
	}

	if o.PresencePenalty != nil {
		req.PresencePenalty = *o.PresencePenalty
	}

	if o.Stop != nil {
		req.Stop = o.Stop
	}

	if o.LogitBias != nil {
		req.LogitBias = o.LogitBias
	}
}
