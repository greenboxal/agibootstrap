package openaiv1

import (
	"encoding/json"
	"io"
	"net/http"
)

type RequestHandler[TReq any] func(req *Request[TReq], writer *ResponseWriter) error

func buildHandler[TReq any](fn RequestHandler[TReq]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &Request[TReq]{Request: r}
		writer := &ResponseWriter{ResponseWriter: w, req: r}

		data, err := io.ReadAll(req.Body)

		if err != nil {
			return
		}

		if len(data) > 0 {
			if err := json.Unmarshal(data, &req.Payload); err != nil {
				writer.WriteError(err)
				return
			}
		}

		if err := fn(req, writer); err != nil {
			writer.WriteError(err)
			return
		}
	}
}

type Request[TReq any] struct {
	*http.Request

	Payload TReq
}

type ResponseWriter struct {
	http.ResponseWriter
	req *http.Request
}

func (w ResponseWriter) WriteResponse(res any) {
	data, err := json.Marshal(res)

	if err != nil {
		w.WriteError(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(data); err != nil {
		panic(err)
	}
}

func (w ResponseWriter) WriteError(err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
