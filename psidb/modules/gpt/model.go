package gpt

import (
	"context"
	"os"
	"path"
	"strings"

	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"
)

var GlobalClient = createNewClient()
var GlobalEmbedder = &openai.Embedder{
	Client: GlobalClient,
	Model:  openai.AdaEmbeddingV2,
}

var GlobalModel = &TracingChatLanguageModel{
	&openai.ChatLanguageModel{
		Client:      GlobalClient,
		Model:       "gpt-3.5-turbo-16k",
		Temperature: 1.5,
	},
}

var GlobalModelTokenizer = tokenizers.TikTokenForModel(GlobalModel.Model)

func createNewClient() *openai.Client {
	if os.Getenv("OPENAI_API_KEY") == "" {
		home := os.Getenv("HOME")

		if home != "" {
			p := path.Join(home, ".openai", "api-key")
			key, err := os.ReadFile(p)

			if err == nil {
				_ = os.Setenv("OPENAI_API_KEY", strings.TrimSpace(string(key)))
			}
		}
	}

	return openai.NewClient()
}

type TracingChatLanguageModel struct {
	*openai.ChatLanguageModel
}

func (lm *TracingChatLanguageModel) PredictChatStream(ctx context.Context, msg chat.Message, options ...llm.PredictOption) (chat.MessageStream, error) {
	opts := llm.NewPredictOptions(options...)
	request, err := lm.BuildChatCompletionRequest(msg, opts)

	if err != nil {
		return nil, err
	}

	trace := CreateTrace(ctx, *request)

	stream, err := lm.Client.CreateChatCompletionStream(ctx, *request)

	if err != nil {
		return nil, err
	}

	return &tracingMessageStream{trace: trace, stream: stream}, nil
}

type tracingMessageStream struct {
	trace  *Trace
	stream *openai.ChatCompletionStream
}

func (m *tracingMessageStream) Recv() (chat.MessageFragment, error) {
	reply, err := m.stream.Recv()

	if err != nil {
		m.trace.ReportError(err)
		m.trace.End()

		return chat.MessageFragment{}, err
	}

	m.trace.ConsumeOpenAI(reply)

	return chat.MessageFragment{
		MessageIndex: reply.Choices[0].Index,
		Delta:        reply.Choices[0].Delta.Content,
	}, nil
}

func (m *tracingMessageStream) Close() error {
	m.trace.End()

	return m.Close()
}
