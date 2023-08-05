package rendering

import (
	"bytes"
	"context"
	"testing"

	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/greenboxal/agibootstrap/pkg/langs/golang"

	"github.com/greenboxal/agibootstrap/pkg/legacy/codex"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func TestPruningRenderer(t *testing.T) {
	var pr *PruningRenderer

	pr = &PruningRenderer{
		Tokenizer: tokenizers.TikTokenForModel(openai.AdaEmbeddingV2.String()),

		Weight: func(state *NodeState, node psi.Node) float32 {
			return 1
		},

		Write: func(w *TokenBuffer, node psi.Node) error {
,,			if node.IsContainer() {
				n, err := w.Write([]byte(node.String() + " {\n"))

				if err != nil {
					return err
				}

				for _, c := range node.Children() {
					_, err = w.WriteNode(pr, c)

					if err != nil {
						return err
					}

					_, err = w.Write([]byte(";\n"))

					if err != nil {
						return err
					}
				}

				_, err = w.Write([]byte("\n}\n"))

				if err != nil {
					return err
				}
			} else {
				_, err := w.Write([]byte(node.String()))

				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	p, err := codex.LoadProject(context.Background(), ".")
	require.NoError(t, err)
	lang := golang.NewLanguage(p)
	src := golang.NewSourceFile(lang, "test.go", repofs.String("package main\n\nfunc main() {\n\tprintln(\"Hello world\")\n}"))
	require.NoError(t, src.Load())
	node := src.Root()

	buf := bytes.NewBuffer(nil)
	_, err = pr.Render(node, buf)
	require.NoError(t, err)

	str := buf.String()

	strs := lo.MapValues(pr.nodeStates, func(s *NodeState, _ string) string {
		return s.Buffer.String()
	})

	require.NotNil(t, strs)

	require.Equal(t, "package main\n\nfunc main() {\n\tprintln(\"Hello world\")\n}", str)
}
