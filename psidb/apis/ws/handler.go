package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/core/pubsub"
)

type Handler struct {
	logger   *otelzap.SugaredLogger
	upgrader websocket.Upgrader

	pubsub         *pubsub.Manager
	sessionManager coreapi.SessionManager
}

func NewHandler(
	pubsub *pubsub.Manager,
	sm coreapi.SessionManager,
) *Handler {
	h := &Handler{
		logger: logging.GetLogger("apis/ws"),

		pubsub:         pubsub,
		sessionManager: sm,
	}

	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,

		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return h
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	conn, err := h.upgrader.Upgrade(writer, request, nil)

	if err != nil {
		log.Println(err)
		return
	}

	client := NewClient(h, conn)

	client.ServeHTTP(writer, request)
}