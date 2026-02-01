package handler

import (
	"fmt"
	"net/http"

	"github.com/Hiroki111/go-backend-example/internal/service"
)

const DefaultPageLimit = 20
const MaxPageLimit = 1000

type Handler struct {
	service *service.Service
}

func NewHandler(
	service *service.Service,
) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
}
