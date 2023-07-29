package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/graphfs"
)

type API struct {
	chi.Router

	vg *graphfs.VirtualGraph
}

func NewAPI(vg *graphfs.VirtualGraph) *API {
	a := &API{
		Router: chi.NewRouter(),
		vg:     vg,
	}

	a.NotFound(a.handleHttp)

	return a
}

func (a *API) handleHttp(writer http.ResponseWriter, request *http.Request) {
	err := func() (err error) {
		switch request.Method {
		case http.MethodGet:
			return a.handleGet(writer, request)

		case http.MethodPut:
			return a.handlePut(writer, request)

		default:
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}()

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	}

	return
}

func (a *API) handleGet(writer http.ResponseWriter, request *http.Request) error {
	ctx := request.Context()
	psiPath, err := psi.ParsePathEx(request.URL.Path, true)

	if err != nil {
		return err
	}

	fn, err := a.vg.Read(ctx, psiPath)

	if err != nil {
		return err
	}

	writer.WriteHeader(http.StatusOK)

	return ipld.EncodeStreaming(writer, typesystem.Wrap(fn), dagjson.Encode)
}

func (a *API) handlePut(writer http.ResponseWriter, request *http.Request) error {

}
