package rendering

import (
	"testing"

	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"
	"github.com/stretchr/testify/require"
)

func TestTokenBuffer(t *testing.T) {
	tb := NewTokenBuffer(tokenizers.TikTokenForModel(openai.AdaEmbeddingV2.String()), 0)

	tb.Write([]byte("Hello "))
	tb.Write([]byte("world"))

	str := tb.String()

	require.Equal(t, "Hello world", str)
}
