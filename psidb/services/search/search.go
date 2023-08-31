package search

import (
	"context"
	"fmt"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi `github.com/greenboxal/agibootstrap/psidb/core/api`
	indexing2 "github.com/greenboxal/agibootstrap/psidb/services/indexing"
)

type SearchRequest struct {
	Graph coreapi.LiveGraph
	Query psi.Node
	Scope psi.Path
	Limit int

	ReturnEmbeddings bool
	ReturnNode       bool
}

type SearchResponse struct {
	psi.NodeBase

	Results []indexing2.NodeSearchHit `json:"results"`
}

var SearchResponseType = psi.DefineNodeType[*SearchResponse]()

type Service struct {
}

func NewSearchService() *Service {
	return &Service{}
}

func (s *Service) Search(ctx context.Context, request *SearchRequest) (*SearchResponse, error) {
	var searchRequest indexing2.SearchRequest

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

	scp, ok := scpNode.(*indexing2.Scope)

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
