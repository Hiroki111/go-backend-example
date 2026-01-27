package cache

import (
	"context"
	"time"
)

type ProductsCacheWarmer interface {
	WarmProduct(ctx context.Context, id uint, ttl time.Duration)
	WarmProductList(ctx context.Context, ttl time.Duration)
}
