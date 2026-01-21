package cache

import (
	"context"
	"time"
)

type NoopProductsCache struct{}

func NewNoopProductsCache() *NoopProductsCache {
	return &NoopProductsCache{}
}

func (c *NoopProductsCache) GetPage(
	ctx context.Context,
	key string,
) (*ProductsPage, bool, error) {
	return nil, false, nil
}

func (c *NoopProductsCache) SetPage(
	ctx context.Context,
	key string,
	page *ProductsPage,
	ttl time.Duration,
) error {
	return nil
}

func (c *NoopProductsCache) InvalidateProducts(
	ctx context.Context,
) error {
	return nil
}
