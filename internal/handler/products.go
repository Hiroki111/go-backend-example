package handler

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Hiroki111/go-backend-example/internal/cache"
	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/repository"
)

const productListTTLWithoutQuery = 60 * time.Minute
const productListTTLWithQuery = 60 * time.Minute
const individualProductTTL = 30 * time.Minute

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	orderBy := r.URL.Query().Get("orderBy")
	sortIn := r.URL.Query().Get("sortIn")
	name := r.URL.Query().Get("name")
	minPrice := r.URL.Query().Get("minPrice")
	maxPrice := r.URL.Query().Get("maxPrice")
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")

	var ttl time.Duration
	isDefaultQuery :=
		orderBy == "" &&
			sortIn == "" &&
			name == "" &&
			minPrice == "" &&
			maxPrice == "" &&
			page == "" &&
			limit == ""
	if isDefaultQuery {
		ttl = productListTTLWithoutQuery
	} else {
		ttl = productListTTLWithQuery
	}
	ttl = addJitter(ttl)

	inputs := getDefaultQueryForProducts()
	inputs.OrderBy = orderBy
	inputs.SortIn = sortIn
	inputs.Name = name

	minPriceInt, err := parseOptionalInt64(minPrice, inputs.MinPrice)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid minPrice",
		})
		return
	}
	inputs.MinPrice = minPriceInt

	maxPriceInt, err := parseOptionalInt64(maxPrice, inputs.MaxPrice)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid maxPrice",
		})
		return
	}
	inputs.MaxPrice = maxPriceInt

	pageInt, err := parseOptionalInt(page, 1)
	if err != nil || pageInt <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "page must be a positive integer",
		})
		return
	}

	limitInt, err := parseOptionalInt(limit, inputs.Limit)
	if err != nil || limitInt <= 0 || limitInt > MaxPageLimit {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "limit must be a positive integer and not exceed " + strconv.Itoa(MaxPageLimit),
		})
		return
	}
	inputs.Limit = limitInt
	inputs.Offset = (pageInt - 1) * inputs.Limit

	ctx := r.Context()
	cacheKey := cache.ProductListCacheKey(inputs)
	productPage, found, err := h.productsCache.GetPage(ctx, cacheKey)
	if err != nil {
		log.Printf("cache read failed: %v", err)
	} else if found {
		writeJSON(w, http.StatusOK, GetProductsResponse{
			Items:   mapProductsToProductItems(productPage.Products),
			Page:    pageInt,
			Limit:   limitInt,
			Total:   int(productPage.Total),
			HasNext: pageInt*limitInt < int(productPage.Total),
		})
		return
	}

	products, total, err := h.repo.GetProductsWithTotalCount(inputs)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "failed to get products",
		})
		return
	}

	newProductPage := cache.ProductsPage{
		Products: products,
		Total:    total,
	}
	if err := h.productsCache.SetPage(ctx, cacheKey, &newProductPage, ttl); err != nil {
		log.Printf("failed to set products cache: %v", err)
	}

	response := GetProductsResponse{
		Items:   mapProductsToProductItems(products),
		Page:    pageInt,
		Limit:   limitInt,
		Total:   int(total),
		HasNext: pageInt*limitInt < int(total),
	}

	writeJSON(w, http.StatusOK, response)
}

// TODO: Create a service layer, and move this function there, so that this function is used for receiving GET /products requests
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

func mapProductsToProductItems(products []domain.Product) []ProductItem {
	items := make([]ProductItem, len(products))
	for i, product := range products {
		items[i] = ProductItem{
			ID:         product.ID,
			Name:       product.Name,
			PriceCents: product.PriceCents,
		}
	}

	return items
}

func (h *Handler) GetProductById(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id64, err := strconv.ParseInt(idString, 10, 64)
	if err != nil || id64 <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid ID",
		})
		return
	}
	id := uint(id64)

	ctx := r.Context()
	cacheKey := cache.ProductCacheKey(id)
	cachedProduct, found, err := h.productsCache.GetProduct(ctx, cacheKey)
	if err != nil {
		log.Printf("cache read failed: %v", err)
	} else if found {
		productItem := ProductItem{
			ID:         cachedProduct.ID,
			Name:       cachedProduct.Name,
			PriceCents: cachedProduct.PriceCents,
		}
		writeJSON(w, http.StatusOK, GetProductResponse{
			Item: productItem,
		})
		return
	}

	product, err := h.repo.GetProductById(id)
	if err != nil {
		if err == repository.ErrItemNotFound {
			writeJSON(w, http.StatusNotFound, ErrorResponse{
				Error: "item not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "internal error",
		})
		return
	}

	if err := h.productsCache.SetProduct(ctx, cacheKey, &product, individualProductTTL); err != nil {
		log.Printf("failed to set product cache: %v", err)
	}

	productItem := ProductItem{
		ID:         product.ID,
		Name:       product.Name,
		PriceCents: product.PriceCents,
	}
	writeJSON(w, http.StatusOK, GetProductResponse{
		Item: productItem,
	})
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var payload CreateProductRequest

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	if payload.Name == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid name",
		})
		return
	}

	product, err := h.repo.CreateProduct(domain.Product{
		Name:       payload.Name,
		PriceCents: payload.PriceCents,
	})

	if err != nil {
		if errors.Is(err, repository.ErrProductAlreadyExists) {
			writeJSON(w, http.StatusConflict, ErrorResponse{
				Error: "duplicate product name",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "internal error",
		})
		return
	}

	go h.productsCacheWarmer.WarmProductList(productListTTLWithoutQuery)
	go h.productsCacheWarmer.WarmProduct(product.ID, individualProductTTL)

	item := ProductItem{
		ID:         product.ID,
		Name:       product.Name,
		PriceCents: product.PriceCents,
	}
	writeJSON(w, http.StatusCreated, CreateProductResponse{
		Item: item,
	})
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id64, err := strconv.ParseInt(idString, 10, 64)
	if err != nil || id64 <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid ID",
		})
		return
	}
	id := uint(id64)

	var payload UpdateProductRequest

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	if payload.Name != nil {
		if strings.TrimSpace(*payload.Name) == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "invalid name",
			})
			return
		}
	}

	product, err := h.repo.UpdateProduct(repository.UpdateProductsInput{ID: id, Name: payload.Name, PriceCents: payload.PriceCents})
	if err != nil {
		if errors.Is(err, repository.ErrItemNotFound) {
			writeJSON(w, http.StatusNotFound, ErrorResponse{
				Error: "item not found",
			})
			return
		}
		if errors.Is(err, repository.ErrProductAlreadyExists) {
			writeJSON(w, http.StatusConflict, ErrorResponse{
				Error: "duplicate product data",
			})
			return
		}
		if errors.Is(err, repository.ErrOptimisticLockFailed) {
			writeJSON(w, http.StatusConflict, ErrorResponse{
				Error: "product data locked",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "internal error",
		})
		return
	}

	go h.productsCacheWarmer.WarmProductList(productListTTLWithoutQuery)
	go h.productsCacheWarmer.WarmProduct(product.ID, individualProductTTL)

	item := ProductItem{
		ID:         product.ID,
		Name:       product.Name,
		PriceCents: product.PriceCents,
	}
	writeJSON(w, http.StatusOK, UpdateProductResponse{Item: item})
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id64, err := strconv.ParseInt(idString, 10, 64)
	if err != nil || id64 <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid ID",
		})
		return
	}
	id := uint(id64)

	err = h.repo.DeleteProduct(id)
	if err != nil {
		if errors.Is(err, repository.ErrItemNotFound) {
			writeJSON(w, http.StatusNotFound, ErrorResponse{
				Error: "product not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "internal error",
		})
		return
	}

	ctx := r.Context()
	if err := h.productsCache.InvalidateProducts(ctx); err != nil {
		log.Printf("cache invalidation failed: %v", err)
	}

	cacheKey := cache.ProductCacheKey(id)
	if err := h.productsCache.InvalidateProduct(ctx, cacheKey); err != nil {
		log.Printf("cache invalidation failed: %v", err)
	}

	writeJSON(w, http.StatusOK, DeleteProductResponse{Message: "success"})
}
