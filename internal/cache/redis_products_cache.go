package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisProductsCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisProductsCache(
	client *redis.Client,
	ttl time.Duration,
) *RedisProductsCache {
	return &RedisProductsCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *RedisProductsCache) GetPage(
	ctx context.Context,
	key string,
) (*ProductsPage, bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

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
) error {
	data, err := json.Marshal(page)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *RedisProductsCache) InvalidateProducts(ctx context.Context) error {
	keys, err := c.client.Keys(ctx, "products:*").Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return c.client.Del(ctx, keys...).Err()
}
