package service

import (
	"context"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/Hiroki111/go-backend-example/internal/cache"
	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/repository"
)

const productListTTLWithoutQuery = 60 * time.Minute
const productListTTLWithQuery = 60 * time.Minute
const individualProductTTL = 30 * time.Minute

type GetProductsParameters struct {
	OrderBy  string
	SortIn   string
	Name     string
	MinPrice uint
	MaxPrice uint
	Page     uint
	Limit    uint
}

func (s *Service) GetProductsWithTotalCount(ctx context.Context, params GetProductsParameters) ([]domain.Product, uint, error) {
	inputs := repository.GetProductsInput{
		OrderBy:  params.OrderBy,
		SortIn:   params.SortIn,
		Name:     params.Name,
		MinPrice: params.MinPrice,
		MaxPrice: params.MaxPrice,
		Limit:    params.Limit,
		Offset:   (params.Page - 1) * params.Limit,
	}
	defaultParams := getDefaultQueryForProducts()

	isDefaultQuery :=
		inputs.OrderBy == defaultParams.OrderBy &&
			inputs.SortIn == defaultParams.SortIn &&
			inputs.Name == defaultParams.Name &&
			inputs.MinPrice == defaultParams.MinPrice &&
			inputs.MaxPrice == defaultParams.MaxPrice &&
			inputs.Offset == defaultParams.Offset &&
			params.Limit == defaultParams.Limit
	var ttl time.Duration
	if isDefaultQuery {
		ttl = productListTTLWithoutQuery
	} else {
		ttl = productListTTLWithQuery
	}
	ttl = addJitter(ttl)

	cacheKey := cache.ProductListCacheKey(inputs)
	productPage, found, err := s.productsCache.GetPage(ctx, cacheKey)
	if err != nil {
		log.Printf("cache read failed: %v", err)
	} else if found {
		return productPage.Products, uint(productPage.Total), nil
	}

	products, total, err := s.repo.GetProductsWithTotalCount(inputs)
	if err != nil {
		return nil, 0, err
	}

	newProductPage := cache.ProductsPage{
		Products: products,
		Total:    total,
	}
	if err := s.productsCache.SetPage(ctx, cacheKey, &newProductPage, ttl); err != nil {
		log.Printf("failed to set products cache: %v", err)
	}

	return products, uint(total), nil
}

func (s *Service) GetProductById(ctx context.Context, id uint) (domain.Product, error) {
	cacheKey := cache.ProductCacheKey(id)
	cachedProduct, found, err := s.productsCache.GetProduct(ctx, cacheKey)
	if err != nil {
		log.Printf("cache read failed: %v", err)
	} else if found {
		return *cachedProduct, nil
	}

	product, err := s.repo.GetProductById(id)
	if err != nil {
		return domain.Product{}, err
	}

	if err := s.productsCache.SetProduct(ctx, cacheKey, &product, individualProductTTL); err != nil {
		log.Printf("failed to set product cache: %v", err)
	}

	return product, nil
}

func (s *Service) CreateProduct(input domain.Product) (domain.Product, error) {
	product, err := s.repo.CreateProduct(domain.Product{
		Name:       input.Name,
		PriceCents: input.PriceCents,
	})

	if err != nil {
		return domain.Product{}, err
	}

	go s.productsCacheWarmer.WarmProductList(productListTTLWithoutQuery)
	go s.productsCacheWarmer.WarmProduct(product.ID, individualProductTTL)

	return product, nil
}

func (s *Service) UpdateProduct(ctx context.Context, input repository.UpdateProductsInput) (domain.Product, error) {
	product, err := s.repo.UpdateProduct(input)
	if err != nil {
		return domain.Product{}, err
	}

	if err := s.productsCache.InvalidateProducts(ctx); err != nil {
		log.Printf("cache invalidation failed: %v", err)
	}
	if err := s.productsCache.SetProduct(ctx, cache.ProductCacheKey(product.ID), &product, individualProductTTL); err != nil {
		log.Printf("cache update failed: %v", err)
	}
	go s.productsCacheWarmer.WarmProductList(productListTTLWithoutQuery)

	return product, nil
}

func (s *Service) DeleteProduct(ctx context.Context, id uint) error {
	err := s.repo.DeleteProduct(id)
	if err != nil {
		return err
	}

	if err := s.productsCache.InvalidateProducts(ctx); err != nil {
		log.Printf("cache invalidation failed: %v", err)
	}

	cacheKey := cache.ProductCacheKey(id)
	if err := s.productsCache.InvalidateProduct(ctx, cacheKey); err != nil {
		log.Printf("cache invalidation failed: %v", err)
	}

	go s.productsCacheWarmer.WarmProductList(productListTTLWithoutQuery)

	return nil
}

// TODO: RedisProductsCacheWarmer uses the same function. Move this somewhere so that it can be shared with RedisProductsCacheWarmer
func getDefaultQueryForProducts() repository.GetProductsInput {
	return repository.GetProductsInput{
		OrderBy:  "",
		SortIn:   "",
		Name:     "",
		MinPrice: 0,
		MaxPrice: math.MaxInt64,
		Offset:   0,
		Limit:    DefaultPageLimit,
	}
}

func addJitter(ttl time.Duration) time.Duration {
	jitter := time.Duration(rand.Int63n(int64(5 * time.Minute)))
	return ttl + jitter
}
