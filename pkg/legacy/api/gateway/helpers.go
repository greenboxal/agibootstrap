package gateway

import (
	"mime"
	"net/http"
	"strings"
)

func Negotiate(
	request *http.Request,
	writer http.ResponseWriter,
	defaultFormat string,
	accepts map[string]func() error,
) error {
	accept := request.Header.Get("Accept")

	if format := request.URL.Query().Get("format"); format != "" {
		switch format {
		case "json":
			accept = "application/json"
		case "cbor":
			accept = "application/dag-cbor"
		default:
			accept = format
		}
	}

	if accept == "" {
		accept = "*/*"
	}

	types := strings.Split(accept, ",")
	types = append(types, defaultFormat)

	for _, t := range types {
		mimeType, _, err := mime.ParseMediaType(t)

		if err != nil {
			continue
		}

		if len(mimeType) == 0 {
			return nil
		}

		if renderer, ok := accepts[mimeType]; ok {
			return renderer()
		}
	}

	writer.WriteHeader(http.StatusNotAcceptable)

	return nil
}
