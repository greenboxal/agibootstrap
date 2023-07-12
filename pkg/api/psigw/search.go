package psigw

import (
	"net/http"
	"strconv"

	"github.com/greenboxal/aip/aip-langchain/pkg/chunkers"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphindex"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
)

type SearchQuery struct {
	psi.NodeBase

	Query string `json:"query,omitempty"`
}

type SearchResultList struct {
	psi.NodeBase

	Hits []graphindex.NodeSearchHit `json:"hit"`
}

var SearchResultListType = psi.DefineNodeType[*SearchResultList]()
var SearchQueryType = psi.DefineNodeType[*SearchQuery]()

func (gw *Gateway) handleSearch(writer http.ResponseWriter, request *http.Request) {
	var err error
	var anchor psi.Node

	q := request.URL.Query().Get("q")
	indexName := request.URL.Query().Get("index")
	limit := uint64(10)

	if str := request.URL.Query().Get("limit"); str != "" {
		limit, err = strconv.ParseUint(str, 10, 64)

		if err != nil {
			http.Error(writer, "invalid limit", http.StatusBadRequest)
			return
		}
	}

	if q == "" {
		http.Error(writer, "missing query", http.StatusBadRequest)
		return
	}

	if indexName == "" {
		http.Error(writer, "missing index", http.StatusBadRequest)
		return
	}

	if anchorPath := request.URL.Query().Get("anchor"); anchorPath != "" {
		p, err := psi.ParsePath(anchorPath)

		if err != nil {
			http.Error(writer, "invalid anchor path", http.StatusBadRequest)
			return
		}

		anchor, err = gw.graph.ResolveNode(request.Context(), p)

		if err != nil {
			http.Error(writer, "invalid anchor path", http.StatusBadRequest)
			return
		}
	} else {
		anchor = gw.graph.Root()
	}

	index, err := gw.indexManager.OpenNodeIndex(request.Context(), indexName, &graphindex.AnchoredEmbedder{
		Base:    gw.project.Embedder(),
		Root:    gw.graph.Root(),
		Anchor:  anchor,
		Chunker: &chunkers.TikToken{},
	})

	if err != nil {
		http.Error(writer, "index not found", http.StatusNotFound)
		return
	}

	defer index.Close()

	queryNode := &SearchQuery{
		Query: q,
	}

	queryNode.Init(queryNode)

	emb, err := index.Embedder().EmbeddingsForNode(request.Context(), queryNode)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if !emb.Next() {
		http.Error(writer, "no embeddings", http.StatusBadRequest)
		return
	}

	results, err := index.Search(request.Context(), graphindex.SearchRequest{
		Query: emb.Value(),
		Limit: int(limit),
	})

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	result := &SearchResultList{}
	result.Init(result)
	result.Hits = iterators.ToSlice(results)

	if err := rendering.RenderNodeResponse(writer, request, ApiTheme, "", result); err != nil {
		logger.Error(err)
	}
}
