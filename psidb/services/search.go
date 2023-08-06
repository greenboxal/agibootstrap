package services

import (
	"context"
	"fmt"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/db/indexing"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
)

type SearchRequest struct {
	Graph *online.LiveGraph
	Query psi.Node
	Scope psi.Path
	Limit int

	ReturnEmbeddings bool
	ReturnNode       bool
}

type SearchResponse struct {
	psi.NodeBase

	Results []indexing.NodeSearchHit `json:"results"`
}

var SearchResponseType = psi.DefineNodeType[*SearchResponse]()

type SearchService struct {
}

func NewSearchService() *SearchService {
	return &SearchService{}
}

func (s *SearchService) Search(ctx context.Context, request *SearchRequest) (*SearchResponse, error) {
	var searchRequest indexing.SearchRequest

	searchRequest.Graph = request.Graph
	searchRequest.Limit = request.Limit
	searchRequest.ReturnNode = request.ReturnNode
	searchRequest.ReturnEmbeddings = request.ReturnEmbeddings

	response := &SearchResponse{}
	response.Init(response, psi.WithNodeType(SearchResponseType))

	scpNode, err := request.Graph.ResolveNode(ctx, request.Scope)

	if err != nil {
		return nil, err
	}

	scp, ok := scpNode.(*indexing.Scope)

	if !ok {
		return nil, fmt.Errorf("scope node is not a scope")
	}

	index, err := scp.GetIndex(ctx)

	if err != nil {
		return nil, err
	}

	it, err := index.Embedder().EmbeddingsForNode(ctx, request.Query)

	if err != nil {
		return nil, err
	}

	if !it.Next() {
		return nil, fmt.Errorf("no embeddings for query")
	}

	searchRequest.Query = it.Value()

	result, err := index.Search(ctx, searchRequest)

	if err != nil {
		return nil, err
	}

	response.Results = iterators.ToSlice(result)

	return response, nil
}
