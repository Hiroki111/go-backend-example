package cache

import (
	"context"
	"time"
)

type NoopProductsCacheWarmer struct {
}

func NewNoopProductsCacheWarmer() *NoopProductsCacheWarmer {
	return &NoopProductsCacheWarmer{}
}

func (w *NoopProductsCacheWarmer) WarmProduct(ctx context.Context, id uint, ttl time.Duration) {}

func (w *NoopProductsCacheWarmer) WarmProductList(ctx context.Context, ttl time.Duration) {}
