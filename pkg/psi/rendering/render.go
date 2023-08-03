package rendering

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
)

func RenderNodeResponse(
	writer http.ResponseWriter,
	request *http.Request,
	theme Theme,
	skinName string,
	node psi.Node,
) error {
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
					Context:  request.Context(),
					Renderer: renderer,
					Buffer:   w,
					Theme:    theme,
					Query:    request.URL.Query(),
				}, node)
			},
		}

		writer.Header().Set("Content-Type", contentType)
		writer.WriteHeader(http.StatusOK)

		tb := NewTokenBuffer(renderer.Tokenizer, 0)

		if err := renderer.Write(tb, node); err != nil {
			return err
		}

		if _, err := tb.WriteTo(writer); err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("no skin found for node type: %s (%T)", node.PsiNodeType(), node)
}

func RenderNodeWithTheme(
	ctx context.Context,
	writer io.Writer,
	theme Theme,
	contentType, skinName string,
	node psi.Node,
) error {
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
				Context:  ctx,
				Query:    url.Values{},
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

type NodeSnapshotEdge struct {
	Key     string        `json:"key"`
	ToIndex int64         `json:"to_index"`
	ToPath  *psi.Path     `json:"to_path"`
	ToLink  ipld.Link     `json:"to_link"`
	ToNode  *NodeSnapshot `json:"to_node,omitempty"`
}

type NodeSnapshot struct {
	ID      int64     `json:"id"`
	Path    psi.Path  `json:"path"`
	Link    ipld.Link `json:"link"`
	Version int64     `json:"version"`

	Edges []NodeSnapshotEdge `json:"edges,omitempty"`
}

type GraphSnapshot struct {
	Nodes map[int64]*NodeSnapshot `json:"nodes"`
}

func buildSnapshot(ctx context.Context, node psi.Node, maxDepth int, nested bool) (gs *GraphSnapshot) {
	var buildNode func(node psi.Node, depth int) *NodeSnapshot

	buildNode = func(node psi.Node, depth int) (ns *NodeSnapshot) {
		snap := node.PsiNodeBase().GetSnapshot()

		if snap == nil {
			return
		}

		if gs.Nodes[snap.ID()] != nil {
			return
		}

		ns = &NodeSnapshot{}
		ns.ID = snap.ID()
		ns.Path = snap.Path()
		ns.Link = snap.CommitLink()
		ns.Version = snap.CommitVersion()

		gs.Nodes[snap.ID()] = ns

		edges, err := node.PsiNodeBase().Graph().ListNodeEdges(ctx, node.CanonicalPath())

		if err != nil {
			return
		}

		for _, fe := range edges {
			es := NodeSnapshotEdge{}

			es.Key = fe.Key.String()
			es.ToIndex = fe.ToIndex
			es.ToPath = fe.ToPath
			es.ToLink = fe.ToLink

			if depth < maxDepth {
				to, err := node.PsiNodeBase().Graph().ResolveNode(ctx, node.CanonicalPath().Child(fe.Key.AsPathElement()))

				if err != nil {
					continue
				}

				n := buildNode(to, depth+1)

				if nested && fe.Key.GetKind() == psi.EdgeKindChild {
					es.ToNode = n
				}
			}

			ns.Edges = append(ns.Edges, es)
		}

		return
	}

	gs = &GraphSnapshot{}
	gs.Nodes = make(map[int64]*NodeSnapshot)

	buildNode(node, 0)

	return
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
		if skin := t.SkinForNode("text/markdown", skinName, node); skin != nil {
			return &markdownToHtmlSkin{SkinRenderer: skin}
		}

		if skin := t.SkinForNode("text/plain", skinName, node); skin != nil {
			return skin
		}

	case "application/json":
		switch skinName {
		case "psi-snapshot":
			return &SkinBase[psi.Node]{
				Name:        skinName,
				ContentType: contentType,
				NodeType:    node.PsiNodeType(),
				RenderFn: func(ctx SkinRendererContext, node psi.Node) error {
					maxDepth := 1
					nested := false

					if d := ctx.Query.Get("nested"); len(d) > 0 {
						nested = d == "true"
					}

					if d := ctx.Query.Get("depth"); len(d) > 0 {
						md, err := strconv.Atoi(d)

						if err != nil {
							return err
						}

						maxDepth = md
					}

					snap := buildSnapshot(ctx.Context, node, maxDepth, nested)

					if snap == nil {
						return fmt.Errorf("no snapshot found for node: %s (%T)", node.PsiNodeType(), node)
					}

					if nested {
						return json.NewEncoder(ctx.Buffer).Encode(snap.Nodes[node.ID()])
					} else {
						return json.NewEncoder(ctx.Buffer).Encode(snap)
					}
				},
			}

		default:
			return &SkinBase[psi.Node]{
				Name:        skinName,
				ContentType: contentType,
				NodeType:    node.PsiNodeType(),
				RenderFn: func(ctx SkinRendererContext, node psi.Node) error {
					return ipld.EncodeStreaming(ctx.Buffer, typesystem.Wrap(node), dagjson.Encode)
				},
			}
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

type markdownToHtmlSkin struct {
	SkinRenderer
}

func (m *markdownToHtmlSkin) GetContentType() string { return "text/html" }
func (m *markdownToHtmlSkin) RenderNode(ctx SkinRendererContext, node psi.Node) error {
	patchedCtx := ctx
	patchedCtx.Buffer = NewTokenBuffer(ctx.Buffer.tokenizer, ctx.Buffer.tokenLimit)

	if err := m.SkinRenderer.RenderNode(patchedCtx, node); err != nil {
		return err
	}

	html := mdutils.MarkdownToHtml(patchedCtx.Buffer.Bytes())

	_, err := ctx.Buffer.Write(html)

	return err
}
