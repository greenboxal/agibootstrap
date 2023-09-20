package restv1

import (
	"encoding/json"
	"net/http"

	"github.com/greenboxal/agibootstrap/psidb/services/typing"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
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
	data, err := json.Marshal(schema)

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
