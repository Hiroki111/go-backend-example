package cache

import "github.com/Hiroki111/go-backend-example/internal/domain"

type ProductsPage struct {
	Products []domain.Product
	Total    int64
}
