package restv1

import (
	"net/http"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"
	"github.com/greenboxal/agibootstrap/psidb/services/typing"
)

type OpenAPISchemaHandler struct {
	manager *typing.Manager
}

func NewOpenAPISchemaHandler(typeManager *typing.Manager) *OpenAPISchemaHandler {
	return &OpenAPISchemaHandler{
		manager: typeManager,
	}
}

func (o *OpenAPISchemaHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	schema := typesystem.Universe().GlobalJsonSchema()
	data, err := schema.MarshalJSON()

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	_, err = writer.Write(data)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}
