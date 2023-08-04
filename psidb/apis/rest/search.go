package rest

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering/themes"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/indexing"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
)

type SearchRequest struct {
	*http.Request

	Graph *online.LiveGraph

	Scope psi.Path
}

type SearchHandler struct {
	core         coreapi.Core
	indexManager *indexing.Manager
}

func NewSearchHandler(
	core coreapi.Core,
	indexManager *indexing.Manager,
) *SearchHandler {
	return &SearchHandler{
		core:         core,
		indexManager: indexManager,
	}
}

func (s *SearchHandler) handleRequest(request *SearchRequest) (psi.Node, error) {
	var searchRequest indexing.SearchRequest

	response := &SearchResponse{}
	response.Init(response, psi.WithNodeType(SearchResponseType))

	if pathStr := request.Request.URL.Query().Get("scope"); pathStr != "" {
		path, err := psi.ParsePath(pathStr)

		if err != nil {
			return nil, err
		}

		request.Scope = path
	}

	if limitStr := request.Request.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)

		if err != nil {
			return nil, err
		}

		searchRequest.Limit = limit
	}

	scpNode, err := request.Graph.ResolveNode(request.Context(), request.Scope)

	if err != nil {
		return nil, err
	}

	scp, ok := scpNode.(*indexing.Scope)

	if !ok {
		return nil, fmt.Errorf("scope node is not a scope")
	}

	index, err := scp.GetIndex(request.Context())

	if err != nil {
		return nil, err
	}

	result, err := index.Search(request.Context(), searchRequest)

	if err != nil {
		return nil, err
	}

	response.Results = iterators.ToSlice(result)

	return response, nil
}

func (s *SearchHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var result psi.Node

	req := &SearchRequest{Request: request}

	defer func() {
		if err := recover(); err != nil {
			s.handleError(writer, req, err)
		}
	}()

	err := s.core.RunTransaction(request.Context(), func(ctx context.Context, txn coreapi.Transaction) error {
		req.Request = req.Request.WithContext(ctx)
		req.Graph = txn.Graph()

		res, err := s.handleRequest(req)

		if err != nil {
			return err
		}

		result = res

		return nil
	})

	if err != nil {
		s.handleError(writer, req, err)
		return
	}

	if err := rendering.RenderNodeResponse(writer, request, themes.GlobalTheme, "", result); err != nil {
		logger.Error(err)
	}
}

func (s *SearchHandler) handleError(writer http.ResponseWriter, req *SearchRequest, e any) {
	err, ok := e.(error)

	if !ok {
		err = fmt.Errorf("%v", e)
	}

	logger.Error(err)

	status := http.StatusInternalServerError

	if httpErr, ok := err.(HttpError); ok {
		status = httpErr.StatusCode()
	} else if err == psi.ErrNodeNotFound {
		status = http.StatusNotFound
	}

	writer.WriteHeader(status)
}
