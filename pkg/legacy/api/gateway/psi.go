package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
)

func (gw *Gateway) handlePsiDb(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	err := func() error {
		path, err := psi.ParsePathEx(request.URL.Path, true)

		if err != nil {
			return err
		}

		if path.IsRelative() {
			path = gw.rootPath.Join(path)
		}

		switch request.Method {
		case http.MethodGet:
			return gw.handlePsiDbGet(ctx, request, writer, path)

		default:
			writer.WriteHeader(http.StatusMethodNotAllowed)

			return nil
		}
	}()

	if err != nil {
		if errors.Is(err, psi.ErrNodeNotFound) {
			writer.WriteHeader(http.StatusNotFound)
		} else {
			logger.Error(err)

			writer.WriteHeader(http.StatusInternalServerError)

			_, _ = writer.Write([]byte(err.Error()))
		}
	}
}

func (gw *Gateway) handlePsiDbGet(
	ctx context.Context,
	request *http.Request,
	writer http.ResponseWriter,
	path psi.Path,
) error {
	n, err := gw.graph.ResolveNode(ctx, path)

	if n == nil || errors.Is(err, psi.ErrNodeNotFound) {
		writer.WriteHeader(http.StatusNotFound)

		return nil
	} else if err != nil {
		return err
	}

	if request.URL.Query().Get("render") == "true" {
		view := request.URL.Query().Get("view")

		return rendering.RenderNodeResponse(writer, request, ApiTheme, view, n)
	}

	return Negotiate(request, writer, "application/json", map[string]func() error{
		"application/json": func() error {
			return ipld.EncodeStreaming(writer, typesystem.Wrap(n), dagjson.Encode)
		},

		"application/dag-cbor": func() error {
			return ipld.EncodeStreaming(writer, typesystem.Wrap(n), dagcbor.Encode)
		},

		"text/plain": func() error {
			porcelain := request.URL.Query().Get("porcelain") == "true"

			edges, err := gw.graph.ListNodeEdges(ctx, path)

			if err != nil {
				return err
			}

			if porcelain {
				for _, e := range edges {
					_, _ = fmt.Fprintf(writer, "%d\t%s\t%s\t%s\n", e.ToIndex, e.Key, e.ToPath, e.ToLink)
				}
			} else {
				t := table.NewWriter()
				t.SetOutputMirror(writer)
				t.AppendHeader(table.Row{"Ino", "Name", "Type", "Path", "Link"})
				for _, e := range edges {
					cn, err := gw.graph.ResolveNode(ctx, path.Child(e.Key.AsPathElement()))

					if err != nil {
						logger.Warn(err)
						continue
					}

					t.AppendRow([]interface{}{e.ToIndex, e.Key, cn.PsiNodeType(), cn.CanonicalPath(), e.ToLink})
				}
				t.AppendSeparator()
				t.Render()
			}

			return nil
		},

		"text/html": func() error {
			edges, err := gw.graph.ListNodeEdges(ctx, path)

			if err != nil {
				return err
			}

			edgeDescriptions := make([]EdgeDescription, len(edges))

			for i, e := range edges {
				cp := path.Child(e.Key.AsPathElement())
				cn, err := gw.graph.ResolveNode(ctx, cp)

				if err != nil {
					logger.Warn(err)
					continue
				}

				edgeDescriptions[i] = EdgeDescription{
					Ino:      e.ToIndex,
					Key:      e.Key,
					ToPath:   cn.CanonicalPath(),
					ToLink:   e.ToLink,
					NodeType: cn.PsiNodeType().Name(),
				}
			}

			return nodeEdgeListTemplate.Execute(writer, struct {
				CurrentPath psi.Path
				Node        psi.Node
				Edges       []EdgeDescription
			}{
				CurrentPath: path,
				Edges:       edgeDescriptions,
				Node:        n,
			})
		},
	})
}
