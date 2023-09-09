package restv1

import (
	"fmt"
	"net/http"

	"github.com/go-errors/errors"
)

type HttpError interface {
	error

	Cause() error
	Format(format string, args ...any) error
	StatusCode() int
	Wrap(err error, skip int) error
	Unwrap() error
}

type httpErrorBase struct {
	error      error
	cause      error
	statusCode int
}

func (h *httpErrorBase) Format(format string, args ...any) error {
	err := errors.Wrap(fmt.Sprintf(format, args...), 1)

	return h.Wrap(err, 1)
}

func (h *httpErrorBase) StatusCode() int { return h.statusCode }

func (h *httpErrorBase) Cause() error { return h.cause }

func (h *httpErrorBase) Unwrap() error {
	if h.cause != nil {
		return h.cause
	}

	return h.error
}

func (h *httpErrorBase) Wrap(err error, skip int) error {
	return &httpErrorBase{
		error:      h.error,
		cause:      errors.Wrap(err, 1+skip),
		statusCode: h.statusCode,
	}
}

func (h *httpErrorBase) Error() string {
	return h.error.Error()
}

func NewHttpError(code int, msg string) HttpError {
	return &httpErrorBase{
		error:      fmt.Errorf("%s", msg),
		statusCode: code,
	}
}

var ErrNotFound = NewHttpError(http.StatusNotFound, "not found")
var ErrBadRequest = NewHttpError(http.StatusBadRequest, "bad request")
var ErrMethodNotAllowed = NewHttpError(http.StatusMethodNotAllowed, "method not allowed")
