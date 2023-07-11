package rendering

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func RenderNodeResponse(writer http.ResponseWriter, request *http.Request, theme Theme, skinName string, node psi.Node) error {
	var renderer *PruningRenderer

	accept := request.Header.Get("Accept")

	if format := request.URL.Query().Get("format"); format != "" {
		switch format {
		case "json":
			accept = "application/json"
		case "cbor":
			accept = "application/dag-cbor"
		case "markdown":
			accept = "text/markdown"
		case "html":
			accept = "text/html"
		default:
			accept = format
		}
	}

	types := strings.Split(accept, ",")

	for _, contentType := range types {
		skin := theme.SkinForNode(contentType, skinName, node)

		if skin == nil {
			skin = getSkinAlternatives(theme, contentType, skinName, node)

			if skin == nil {
				continue
			}
		}

		renderer = &PruningRenderer{
			Tokenizer: gpt.GlobalModelTokenizer,

			Weight: func(state *NodeState, node psi.Node) float32 {
				return 1
			},

			Write: func(w *TokenBuffer, node psi.Node) error {
				skin := theme.SkinForNode(contentType, "", node)

				if skin == nil {
					skin = getSkinAlternatives(theme, contentType, skinName, node)

					if skin == nil {
						return fmt.Errorf("no skin found for node type: %s (%T)", node.PsiNodeType(), node)
					}
				}

				return skin.RenderNode(SkinRendererContext{
					Renderer: renderer,
					Buffer:   w,
					Theme:    theme,
				}, node)
			},
		}

		writer.Header().Set("Content-Type", contentType)
		writer.WriteHeader(http.StatusOK)

		_, err := renderer.Render(node, writer)

		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("no skin found for node type: %s (%T)", node.PsiNodeType(), node)
}

func RenderNodeWithTheme(writer io.Writer, theme Theme, contentType, skinName string, node psi.Node) error {
	var renderer *PruningRenderer

	skin := theme.SkinForNode(contentType, skinName, node)

	if skin == nil {
		skin = getSkinAlternatives(theme, contentType, skinName, node)

		if skin == nil {
			return fmt.Errorf("no skin found for node type: %s (%T)", node.PsiNodeType(), node)
		}
	}

	renderer = &PruningRenderer{
		Tokenizer: gpt.GlobalModelTokenizer,

		Weight: func(state *NodeState, node psi.Node) float32 {
			return 1
		},

		Write: func(w *TokenBuffer, node psi.Node) error {
			skin := theme.SkinForNode(contentType, "", node)

			if skin == nil {
				skin = getSkinAlternatives(theme, contentType, skinName, node)

				if skin == nil {
					return fmt.Errorf("no skin found for node type: %s (%T)", node.PsiNodeType(), node)
				}
			}

			return skin.RenderNode(SkinRendererContext{
				Renderer: renderer,
				Buffer:   w,
				Theme:    theme,
			}, node)
		},
	}

	_, err := renderer.Render(node, writer)

	if err != nil {
		return err
	}

	return nil
}

func getSkinAlternatives(t Theme, contentType string, skinName string, node psi.Node) SkinRenderer {
	switch contentType {
	case "text/plain":
		if skin := t.SkinForNode("text/markdown", skinName, node); skin != nil {
			return skin
		}

	case "text/markdown":
		if skin := t.SkinForNode("text/plain", skinName, node); skin != nil {
			return skin
		}

	case "text/html":
		// TODO: Wrap and render markdown to HTML
		if skin := t.SkinForNode("text/markdown", skinName, node); skin != nil {
			return skin
		}

		if skin := t.SkinForNode("text/plain", skinName, node); skin != nil {
			return skin
		}

	case "application/json":
		return &SkinBase[psi.Node]{
			Name:        skinName,
			ContentType: contentType,
			NodeType:    node.PsiNodeType(),
			RenderFn: func(ctx SkinRendererContext, node psi.Node) error {
				return ipld.EncodeStreaming(ctx.Buffer, typesystem.Wrap(node), dagjson.Encode)
			},
		}
	}

	return &SkinBase[psi.Node]{
		Name:        skinName,
		ContentType: contentType,
		NodeType:    node.PsiNodeType(),
		RenderFn: func(ctx SkinRendererContext, node psi.Node) error {
			data, err := ipld.Encode(typesystem.Wrap(node), dagjson.Encode)

			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(ctx.Buffer, "%s (%T) %s = %s\n", node.PsiNodeType(), node, data, node)

			if err != nil {
				return err
			}

			return nil
		},
	}
}
