package gpt

import (
	"context"

	"github.com/google/uuid"
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/hashicorp/go-multierror"
	openai2 "github.com/sashabaranov/go-openai"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type StreamingTraceChunk struct {
	TraceID      string              `json:"trace_id"`
	Index        int                 `json:"choice_index"`
	FinishReason openai.FinishReason `json:"finish_reason"`

	Role         msn.Role             `json:"role"`
	FunctionCall *openai.FunctionCall `json:"function_call"`
	Content      string               `json:"content"`

	Tags []string `json:"tags"`
}

type Trace struct {
	sess coreapi.Session

	TraceID  string                         `json:"trace_id"`
	Messages []openai.ChatCompletionMessage `json:"messages"`
	Choices  []openai.ChatCompletionChoice  `json:"choices"`
	Error    error                          `json:"error"`
	Done     bool                           `json:"done"`
	Tags     []string                       `json:"tags"`
}

func CreateTrace(ctx context.Context, req openai.ChatCompletionRequest, tags ...string) *Trace {
	sess := coreapi.GetSession(ctx)

	n := req.N

	if n < 1 {
		n = 1
	}

	t := &Trace{
		sess: sess,

		TraceID:  uuid.NewString(),
		Messages: req.Messages,
		Choices:  make([]openai.ChatCompletionChoice, n),
		Tags:     append(tags, psi.GetTraceTags(ctx)...),
	}

	for i := 0; i < len(t.Choices); i++ {
		t.dispatchChunk(StreamingTraceChunk{
			TraceID:      t.TraceID,
			Index:        i,
			FinishReason: "",
			Content:      "",
			FunctionCall: nil,
			Tags:         t.Tags,
		})
	}

	return t
}

func (t *Trace) ConsumeChunk(chunk StreamingTraceChunk) {
	if len(t.Choices) <= chunk.Index {
		n := chunk.Index + 1

		choices := make([]openai.ChatCompletionChoice, n)
		copy(choices, t.Choices)
		t.Choices = choices
	}

	choice := &t.Choices[chunk.Index]

	choice.Index = chunk.Index
	choice.Message.Content += chunk.Content
	choice.FinishReason = chunk.FinishReason

	if chunk.Role != "" {
		choice.Message.Role = openai.ConvertFromRole(chunk.Role)
	}

	if chunk.FunctionCall != nil {
		if choice.Message.FunctionCall == nil {
			choice.Message.FunctionCall = &openai.FunctionCall{}
		}

		if chunk.FunctionCall.Name != "" {
			choice.Message.FunctionCall.Name = chunk.FunctionCall.Name
		}

		choice.Message.FunctionCall.Arguments += chunk.FunctionCall.Arguments
	}

	t.dispatchChunk(chunk)
}

func (t *Trace) ReportError(err error) {
	if t.Error != nil {
		t.Error = multierror.Append(t.Error, err)
	} else {
		t.Error = err
	}
}

func (t *Trace) End() {
	t.Done = true

	t.onTraceFinished()
}

func (t *Trace) dispatchChunk(chunk StreamingTraceChunk) {
	if t.sess == nil {
		return
	}

	chunk.TraceID = t.TraceID

	t.sess.SendMessage(&SessionMessageGPTraceChunk{
		Chunk: chunk,
	})
}

func (t *Trace) onTraceFinished() {
	if t.sess == nil {
		return
	}

	t.sess.SendMessage(&SessionMessageGPTrace{
		Trace: *t,
	})
}

func (t *Trace) ConsumeOpenAI(chunk openai2.ChatCompletionStreamResponse) {
	for _, choice := range chunk.Choices {
		ck := StreamingTraceChunk{
			TraceID:      t.TraceID,
			Index:        choice.Index,
			FinishReason: choice.FinishReason,
			Content:      choice.Delta.Content,
			FunctionCall: choice.Delta.FunctionCall,
		}

		if choice.Delta.Role != "" {
			ck.Role = openai.ConvertToRole(choice.Delta.Role)
		}

		t.ConsumeChunk(ck)
	}
}

type SessionMessageGPTraceChunk struct {
	coreapi.SessionMessageBase

	Chunk StreamingTraceChunk `json:"chunk"`
}
type SessionMessageGPTrace struct {
	coreapi.SessionMessageBase

	Trace Trace `json:"trace"`
}

func init() {
	typesystem.GetType[SessionMessageGPTrace]()
	typesystem.GetType[SessionMessageGPTraceChunk]()
}
