package promptml

import (
	"bytes"
	"context"
	"testing"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"
	"github.com/stretchr/testify/require"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

func TestRenderingSimple(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := Container(
		Message("me", "you", msn.RoleUser, Text("Hello!")),
		Message("you", "me", msn.RoleAI, Text("Hello, world!")),
	)

	tokenizer := tokenizers.TikTokenForModel("gpt-3.5-turbo")
	stage := NewStage(root, tokenizer)

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

	tokenizer := tokenizers.TikTokenForModel("gpt-3.5-turbo")
	stage := NewStage(root, tokenizer)

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
