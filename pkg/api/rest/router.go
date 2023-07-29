package rest

import (
	"fmt"
	"io"
	"net/http"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	contentnegotiation "gitlab.com/jamietanna/content-negotiation-go"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/online"
)

var logger = logging.GetLogger("api/rest")

type Request struct {
	*http.Request

	ContentType     *contentnegotiation.MediaType
	AcceptedFormats []contentnegotiation.MediaType

	PsiPath psi.Path
}

type Router struct {
	lg *online.LiveGraph
}

func NewRouter(lg *online.LiveGraph) *Router {
	return &Router{
		lg: lg,
	}
}

func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	req := &Request{Request: request}

	defer func() {
		if err := recover(); err != nil {
			r.handleError(writer, req, err)
		}
	}()

	res, err := r.handleRequest(writer, req)

	if err != nil {
		r.handleError(writer, req, err)
		return
	}

	writer.WriteHeader(http.StatusOK)

	if res != nil {
		if err := ipld.EncodeStreaming(writer, typesystem.Wrap(res), dagjson.Encode); err != nil {
			logger.Error(err)
		}
	}
}

func (r *Router) handleRequest(writer http.ResponseWriter, req *Request) (any, error) {
	req.AcceptedFormats = contentnegotiation.ParseAcceptHeaders(req.Header.Values("Accept")...)

	if s := req.Header.Get("Content-Type"); s != "" {
		req.ContentType = contentnegotiation.NewMediaType(s)
	}

	parsedPath, err := psi.ParsePathEx(req.URL.Path, true)

	if err != nil {
		return nil, err
	}

	req.PsiPath = parsedPath

	switch req.Method {
	case http.MethodHead:
		fallthrough
	case http.MethodGet:
		return r.handleGet(req)

	case http.MethodPost:
		return r.handlePost(req)

	case http.MethodPut:
		return r.handlePut(req)

	case http.MethodPatch:
		return r.handlePatch(req)

	case http.MethodDelete:
		return r.handleDelete(req)
	}

	return nil, ErrMethodNotAllowed
}

func (r *Router) handleGet(request *Request) (any, error) {
	return r.lg.ResolveNode(request.Context(), request.PsiPath)
}

func (r *Router) handlePost(request *Request) (any, error) {
	var dataReader io.Reader
	var nodeType psi.NodeType

	if request.ContentType.GetType() == "multipart" && request.ContentType.GetSubType() == "form-data" {
		if err := request.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}

		if s := request.FormValue("type"); s != "" {
			nodeType = psi.NodeTypeByName(s)
		} else {
			nodeType = vfs.FileType
		}

		file, _, err := request.FormFile("data")

		if err != nil {
			return nil, err
		}

		dataReader = file
	} else {
		return nil, ErrMethodNotAllowed
	}

	if nodeType == nil {
		return nil, NewHttpError(http.StatusBadRequest, "invalid node type")
	}

	if nodeType == vfs.FileType {

	} else {
		panic("not implemented")
	}

	return nil, nil
}

func (r *Router) handlePut(request *Request) (any, error) {
	return nil, nil
}

func (r *Router) handlePatch(request *Request) (any, error) {
	return nil, nil
}

func (r *Router) handleDelete(request *Request) (any, error) {
	node, err := r.lg.ResolveNode(request.Context(), request.PsiPath)

	if err != nil {
		return nil, err
	}

	if err := r.lg.Remove(request.Context(), node); err != nil {
		return nil, err
	}

	return node, nil
}

func (r *Router) handleError(writer http.ResponseWriter, request *Request, e any) {
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
