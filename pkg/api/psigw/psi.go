package psigw

import (
	"context"
	"fmt"
	"net/http"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
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
		if err == psi.ErrNodeNotFound {
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

	if n == nil || err == psi.ErrNodeNotFound {
		writer.WriteHeader(http.StatusNotFound)

		return nil
	} else if err != nil {
		return err
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

			edges, err := gw.graph.Store().ListNodeEdges(ctx, n.CanonicalPath())

			if err != nil {
				return err
			}

			if porcelain {
				for edges.Next() {
					e := edges.Value()

					_, _ = fmt.Fprintf(writer, "%d\t%s\t%s\t%s\n", e.ToIndex, e.Key, e.ToPath, e.ToLink)
				}
			} else {
				t := table.NewWriter()
				t.SetOutputMirror(writer)
				t.AppendHeader(table.Row{"Ino", "Name", "Path", "Link"})
				for edges.Next() {
					e := edges.Value()

					t.AppendRow([]interface{}{e.ToIndex, e.Key, e.ToPath, e.ToLink})
				}
				t.AppendSeparator()
				t.Render()
			}

			return nil
		},

		"text/html": func() error {
			edges, err := gw.graph.Store().ListNodeEdges(ctx, n.CanonicalPath())

			if err != nil {
				return err
			}

			edgeList := iterators.ToSlice(edges)

			return nodeEdgeListTemplate.Execute(writer, struct {
				CurrentPath psi.Path
				Node        psi.Node
				Edges       []*psi.FrozenEdge
			}{
				CurrentPath: path,
				Edges:       edgeList,
				Node:        n,
			})
		},
	})
}
