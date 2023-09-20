package autoform

import (
	"context"
	"testing"

	"github.com/alecthomas/repr"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/stretchr/testify/require"
)

type TestFormTask1 struct {
	Title      string   `json:"title"`
	Subtitle   string   `json:"subtitle"`
	Categories []string `json:"categories"`
}

func TestFormTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	history := MessageSourceFromModelMessages(
		openai.ChatCompletionMessage{
			Role:    "user",
			Content: `Who is Georgianna Hopley?`,
		},
		openai.ChatCompletionMessage{
			Role:    "assistant",
			Content: `Georgianna Eliza Hopley (1858–1944) was an American journalist, political figure, and temperance advocate. A member of a prominent Ohio publishing family, she was the first woman reporter in Columbus, and editor of several publications. She served as a correspondent and representative at the 1900 Paris Exposition and the 1901 Pan-American Exposition. She was active in state and national politics, serving as vice-president of the Woman's Republican Club of Ohio and directing publicity for Warren G. Harding's presidential campaign. In 1922 Hopley became the first woman prohibition agent of the United States Bureau of Prohibition, where she was involved in education and publicity.`,
		},
	)

	f := NewForm(WithSchemaFor[TestFormTask1]())
	ft := NewFormTask(f, history)
	ok, err := ft.RunToCompletion(ctx)

	require.NoError(t, err)
	require.True(t, ok)

	result, err := ft.form.Build()

	require.NoError(t, err)

	repr.Println(result)
}
