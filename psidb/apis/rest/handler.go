package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	contentnegotiation "gitlab.com/jamietanna/content-negotiation-go"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering/themes"
	"github.com/greenboxal/agibootstrap/psidb/core"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
)

var logger = logging.GetLogger("api/rest")

var decoderMap = map[string]ipld.Decoder{
	"application/json": dagjson.Decode,
	"application/cbor": dagcbor.Decode,
}

type Request struct {
	*http.Request

	Graph *online.LiveGraph

	ContentType     *contentnegotiation.MediaType
	AcceptedFormats []contentnegotiation.MediaType

	PsiPath psi.Path
}

type Handler struct {
	core *core.Core
}

func NewRouter(core *core.Core) *Handler {
	return &Handler{
		core: core,
	}
}

func (r *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var result any

	req := &Request{Request: request}

	defer func() {
		if err := recover(); err != nil {
			r.handleError(writer, req, err)
		}
	}()

	err := r.core.RunTransaction(request.Context(), func(ctx context.Context, txn coreapi.Transaction) error {
		req.Request = req.Request.WithContext(ctx)
		req.Graph = txn.Graph()

		res, err := r.handleRequest(req)

		if err != nil {
			return err
		}

		result = res

		return nil
	})

	if err != nil {
		r.handleError(writer, req, err)
		return
	}

	if n, ok := result.(psi.Node); ok {
		writer.Header().Set("X-Psi-Path", n.CanonicalPath().String())
		writer.Header().Set("X-Psi-Node-Type", n.PsiNodeType().Name())
		writer.Header().Set("X-Psi-Node-Version", strconv.FormatInt(n.PsiNodeVersion(), 10))
		writer.Header().Set("X-Psi-Node-Index", strconv.FormatInt(n.ID(), 10))
	}

	writer.WriteHeader(http.StatusOK)

	if result != nil {
		view := request.URL.Query().Get("view")

		if request.Method != http.MethodHead {
			if n, ok := result.(psi.Node); ok {
				if request.Header.Get("Accept") == "" {
					request.Header.Set("Accept", "application/json")
				}

				err := rendering.RenderNodeResponse(writer, request, themes.GlobalTheme, view, n)

				if err != nil {
					logger.Error(err)
				}
			} else {
				writer.Header().Set("Content-Type", "application/json")

				if err := ipld.EncodeStreaming(writer, typesystem.Wrap(result), dagjson.Encode); err != nil {
					logger.Error(err)
				}
			}
		}
	}
}

func (r *Handler) handleRequest(req *Request) (any, error) {
	req.AcceptedFormats = contentnegotiation.ParseAcceptHeaders(req.Header.Values("Accept")...)

	if s := req.Header.Get("Content-Type"); s != "" {
		req.ContentType = contentnegotiation.NewMediaType(s)
	}

	pathStr := req.URL.Path
	pathStr = strings.TrimPrefix(pathStr, "/")
	parsedPath, err := psi.ParsePathEx(pathStr, true)

	if err != nil {
		return nil, err
	}

	if parsedPath.IsRelative() {
		parsedPath = parsedPath.WithRoot(r.core.Config().RootUUID)
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

func (r *Handler) handleGet(request *Request) (any, error) {
	return request.Graph.ResolveNode(request.Context(), request.PsiPath)
}

func (r *Handler) handlePost(request *Request) (any, error) {
	var dataReader io.Reader
	var nodeType psi.NodeType

	if s := request.FormValue("type"); s != "" {
		nodeType = psi.NodeTypeByName(s)
	} else if s := request.URL.Query().Get("type"); s != "" {
		nodeType = psi.NodeTypeByName(s)
	} else if s := request.Header.Get("X-Psi-Node-Type"); s != "" {
		nodeType = psi.NodeTypeByName(s)
	}

	if request.ContentType == nil {
		return nil, ErrBadRequest
	}

	if request.ContentType.GetType() == "multipart" && request.ContentType.GetSubType() == "form-data" {
		if err := request.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}

		if nodeType == nil {
			nodeType = vfs.FileType
		}

		file, _, err := request.FormFile("data")

		if err != nil {
			return nil, err
		}

		dataReader = file
	} else {
		dataReader = request.Body
	}

	if nodeType == nil {
		return nil, NewHttpError(http.StatusBadRequest, "invalid node type")
	}

	if nodeType == vfs.FileType {
		return r.handlePostFile(request, dataReader)
	}

	decoder := decoderMap[request.ContentType.String()]

	if decoder == nil {
		return nil, NewHttpError(http.StatusBadRequest, "invalid content type")
	}

	wrapped, err := ipld.DecodeStreamingUsingPrototype(dataReader, decoder, nodeType.Type().IpldPrototype())

	if err != nil {
		return nil, err
	}

	node, ok := typesystem.TryUnwrap[psi.Node](wrapped)

	if !ok {
		return nil, NewHttpError(http.StatusBadRequest, "invalid node type")
	}

	nodeType.InitializeNode(node)

	parentPath := request.PsiPath.Parent()
	parent, err := request.Graph.ResolveNode(request.Context(), parentPath)

	if err != nil {
		return nil, err
	}

	node.SetParent(parent)
	parent.SetEdge(request.PsiPath.Name().GetKey(), node)

	if err := parent.Update(request.Context()); err != nil {
		return nil, err
	}

	return node, nil
}

func (r *Handler) handlePut(request *Request) (any, error) {
	return nil, nil
}

func (r *Handler) handlePatch(request *Request) (any, error) {
	return nil, nil
}

func (r *Handler) handleDelete(request *Request) (any, error) {
	node, err := request.Graph.ResolveNode(request.Context(), request.PsiPath)

	if err != nil {
		return nil, err
	}

	if err := request.Graph.Delete(request.Context(), node); err != nil {
		return nil, err
	}

	return node, nil
}

func (r *Handler) handleError(writer http.ResponseWriter, request *Request, e any) {
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

func (r *Handler) handlePostFile(request *Request, reader io.Reader) (any, error) {
	f, err := request.Graph.ResolveNode(request.Context(), request.PsiPath)

	if err == psi.ErrNodeNotFound {
		parentPath := request.PsiPath.Parent()
		parent, err := request.Graph.ResolveNode(request.Context(), parentPath)

		if err != nil {
			return nil, err
		}

		d, ok := parent.(*vfs.Directory)

		if !ok {
			return nil, fmt.Errorf("parent node is not a directory")
		}

		f, err = d.GetOrCreateFile(request.Context(), request.PsiPath.Name().Name)

		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	if f == nil {
		return nil, ErrNotFound
	}

	vf := f.(*vfs.File)
	fh, err := vf.Open()

	if err != nil {
		return nil, err
	}

	defer fh.Close()

	if err := fh.Put(reader); err != nil {
		return nil, err
	}

	if err := f.Update(request.Context()); err != nil {
		return nil, err
	}

	return f, nil
}