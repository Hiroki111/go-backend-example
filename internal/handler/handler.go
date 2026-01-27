package handler

import (
	"fmt"
	"net/http"

	"github.com/Hiroki111/go-backend-example/internal/cache"
	"github.com/Hiroki111/go-backend-example/internal/repository"
)

const DefaultPageLimit = 20
const MaxPageLimit = 1000

type Handler struct {
	repo                *repository.Repository
	productsCache       cache.ProductsCache
	productsCacheWarmer cache.ProductsCacheWarmer
}

func NewHandler(
	repo *repository.Repository,
	productsCache cache.ProductsCache,
	productsCacheWarmer cache.ProductsCacheWarmer,
) *Handler {
	return &Handler{
		repo:                repo,
		productsCache:       productsCache,
		productsCacheWarmer: productsCacheWarmer,
	}
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
}
