package cache

import (
	"context"
	"time"
)

type ProductsCache interface {
	GetPage(ctx context.Context, key string) (*ProductsPage, bool, error)
	SetPage(ctx context.Context, key string, page *ProductsPage, ttl time.Duration) error
	InvalidateProducts(ctx context.Context) error
}
