package promptml

import (
	"bytes"
	"context"
	"testing"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

var testTokenizer = tokenizers.TikTokenForModel(openai.GPT4)

func TestRenderingSimple(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := Container(
		Message("me", msn.RoleUser, Text("Hello!")),
		Message("you", msn.RoleAI, Text("Hello, world!")),
	)

	stage := NewStage(root, testTokenizer)
	stage.MaxTokens = 10

	buf := bytes.NewBuffer(nil)

	if err := stage.Render(ctx, buf); err != nil {
		t.Fatal(err)
	}

	str := buf.String()

	require.NotEmpty(t, str)
}

func TestRenderingBinding(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	roleUser := obsfx.Just(msn.RoleUser)
	roleAI := obsfx.Just(msn.RoleAI)
	me := obsfx.Just("me")
	you := obsfx.Just("you")
	text1 := obsfx.StringProperty{}
	text2 := obsfx.StringProperty{}

	text1.SetValue("Hello\n")
	text2.SetValue("Hi\n")

	root := Container(
		MessageWithData(me, you, roleUser, TextWithData(&text1)),
		MessageWithData(you, me, roleAI, TextWithData(&text2)),
	)

	stage := NewStage(root, testTokenizer)
	stage.MaxTokens = 15

	str1, err := stage.RenderToString(ctx)

	require.NoError(t, err)
	require.NotEmpty(t, str1)

	text1.SetValue("Other value\n")
	text2.SetValue("Yes, other value\n")

	str2, err := stage.RenderToString(ctx)

	require.NoError(t, err)
	require.NotEmpty(t, str2)

	require.NotEqual(t, str1, str2)
}

func TestRenderingDynamicList(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := Container(
		Message("System", msn.RoleSystem, Container(
			MakeFixed(Text("Hello!")),
			MakeFixed(Text("Available Context:")),

			NewDynamicList(func(ctx context.Context) iterators.Iterator[Node] {
				return iterators.FromSlice([]Node{
					Text("Hello"),
					Text("World"),
				})
			}),

			MakeFixed(Text("End of Available Context.")),
		)),

		Message("Human", msn.RoleUser, MakeFixed(Text("Say something about the context above."))),
		Message("AI", msn.RoleAI, MakeFixed(Text(" "))),
	)

	stage := NewStage(root, testTokenizer)
	stage.MaxTokens = 7

	str1, err := stage.RenderToString(ctx)

	require.NoError(t, err)
	require.Equal(t, "Hello!Available Context:HelloWorldEnd of Available Context.Say something about the context above. ", str1)
}
