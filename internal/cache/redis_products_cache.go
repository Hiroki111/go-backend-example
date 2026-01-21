package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Hiroki111/go-backend-example/internal/metrics"
	"github.com/redis/go-redis/v9"
)

type RedisProductsCache struct {
	client *redis.Client
}

func NewRedisProductsCache(
	client *redis.Client,
) *RedisProductsCache {
	return &RedisProductsCache{
		client: client,
	}
}

func (c *RedisProductsCache) GetPage(
	ctx context.Context,
	key string,
) (*ProductsPage, bool, error) {
	val, err := c.client.Get(ctx, key).Result()

	if err == redis.Nil {
		metrics.ProductsCacheMisses.Inc()
		return nil, false, ErrCacheMiss
	}

	if err != nil {
		return nil, false, err
	}

	metrics.ProductsCacheHits.Inc()

	var products ProductsPage
	if err := json.Unmarshal([]byte(val), &products); err != nil {
		return nil, false, err
	}

	return &products, true, nil
}

func (c *RedisProductsCache) SetPage(
	ctx context.Context,
	key string,
	page *ProductsPage,
	ttl time.Duration,
) error {
	data, err := json.Marshal(page)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisProductsCache) InvalidateProducts(ctx context.Context) error {
	iter := c.client.Scan(ctx, 0, "products:*", 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
