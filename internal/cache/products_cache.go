package cache

import (
	"context"
)

type ProductsCache interface {
	GetPage(ctx context.Context, key string) (*ProductsPage, bool, error)
	SetPage(ctx context.Context, key string, page *ProductsPage) error
	InvalidateProducts(ctx context.Context) error
}
