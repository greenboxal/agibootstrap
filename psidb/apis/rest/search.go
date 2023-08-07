package rest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering/themes"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/services/search"
)

type SearchRequest struct {
	*http.Request

	Graph *online.LiveGraph
}

type SearchHandler struct {
	core   coreapi.Core
	search *search.SearchService
}

func NewSearchHandler(
	core coreapi.Core,
	search *search.SearchService,
) *SearchHandler {
	return &SearchHandler{
		core:   core,
		search: search,
	}
}

func (s *SearchHandler) handleRequest(request *SearchRequest) (psi.Node, error) {
	var searchRequest search.SearchRequest

	searchRequest.Graph = request.Graph
	searchRequest.Limit = 10
	searchRequest.ReturnNode = true

	if pathStr := request.Request.URL.Query().Get("scope"); pathStr != "" {
		pathStr, err := url.QueryUnescape(pathStr)

		if err != nil {
			return nil, err
		}

		path, err := psi.ParsePath(pathStr)

		if err != nil {
			return nil, err
		}

		searchRequest.Scope = path
	}

	if limitStr := request.Request.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)

		if err != nil {
			return nil, err
		}

		searchRequest.Limit = limit
	}

	if queryStr := request.Request.URL.Query().Get("query"); queryStr != "" {
		queryNode := stdlib.NewText(queryStr)
		searchRequest.Query = queryNode
	}

	result, err := s.search.Search(request.Context(), &searchRequest)

	if err != nil {
		return nil, err
	}

	return result, nil
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
