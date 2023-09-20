package markdown

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
)

const TestInput = `
# Heading 1
## Heading 22

Text.

Text.
Block.

With with link: [link](https://example.com) and [link2](https://example.com) and more.
And a block.

And a codeblock:
` + "```" + `
With code.
` + "```" + `

` + "```" + `json
{
	"key": "value",
}
` + "```" + `
`

func TestParser(t *testing.T) {
	p := sparsing.NewParserStream()
	p.PushInStack(&Container{})

	n, err := p.Write([]byte(TestInput))

	require.NoError(t, err)
	require.Equal(t, len(TestInput), n)

	err = p.Close()

	require.NoError(t, err)
}
