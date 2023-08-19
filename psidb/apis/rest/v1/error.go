package restv1

import (
	"fmt"
	"net/http"
)

type HttpError interface {
	error

	StatusCode() int
}

type httpErrorBase struct {
	error
	statusCode int
}

func (h httpErrorBase) StatusCode() int { return h.statusCode }

func NewHttpError(code int, msg string) HttpError {
	return httpErrorBase{
		error:      fmt.Errorf(msg),
		statusCode: code,
	}
}

var ErrNotFound = NewHttpError(http.StatusNotFound, "not found")
var ErrBadRequest = NewHttpError(http.StatusBadRequest, "bad request")
var ErrMethodNotAllowed = NewHttpError(http.StatusMethodNotAllowed, "method not allowed")
