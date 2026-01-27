package cache

import (
	"context"
	"log"
	"time"

	"github.com/Hiroki111/go-backend-example/internal/repository"
)

type RedisProductsCacheWarmer struct {
	repo    repository.Repository
	cache   ProductsCache
	workers chan struct{}
}

func NewRedisProductsCacheWarmer(repo repository.Repository, cache ProductsCache) *RedisProductsCacheWarmer {
	return &RedisProductsCacheWarmer{
		repo:    repo,
		cache:   cache,
		workers: make(chan struct{}, 5),
	}
}

func (w *RedisProductsCacheWarmer) WarmProduct(ctx context.Context, id uint, ttl time.Duration) {
	select {
	case w.workers <- struct{}{}:
		go func() {
			defer func() { <-w.workers }()

			product, err := w.repo.GetProductById(id)
			if err != nil {
				log.Printf("failed to get product for cache: %v", err)
				return
			}

			cacheKey := ProductCacheKey(id)
			err = w.cache.SetProduct(ctx, cacheKey, &product, ttl)
			if err != nil {
				log.Printf("failed to warm product cache: %v", err)
				return
			}
		}()
	default:
	}
}

func (w *RedisProductsCacheWarmer) WarmProductList(ctx context.Context, ttl time.Duration) {
	select {
	case w.workers <- struct{}{}:
		go func() {
			defer func() { <-w.workers }()

			inputs := repository.GetProductsInput{}
			products, total, err := w.repo.GetProductsWithTotalCount(inputs)
			if err != nil {
				log.Printf("failed to get products and total count for cache: %v", err)
				return
			}

			cacheKey := ProductListCacheKey(inputs)
			page := ProductsPage{Products: products, Total: total}
			err = w.cache.SetPage(ctx, cacheKey, &page, ttl)
			if err != nil {
				log.Printf("failed to warm product page cache: %v", err)
				return
			}
		}()
	default:
	}
}
