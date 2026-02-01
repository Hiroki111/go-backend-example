package service

import (
	"github.com/Hiroki111/go-backend-example/internal/cache"
	"github.com/Hiroki111/go-backend-example/internal/repository"
)

const DefaultPageLimit = 20
const MaxPageLimit = 1000

type Service struct {
	repo                *repository.Repository
	productsCache       cache.ProductsCache
	productsCacheWarmer cache.ProductsCacheWarmer
}

func NewService(
	repo *repository.Repository,
	productsCache cache.ProductsCache,
	productsCacheWarmer cache.ProductsCacheWarmer,
) *Service {
	return &Service{
		repo:                repo,
		productsCache:       productsCache,
		productsCacheWarmer: productsCacheWarmer,
	}
}
