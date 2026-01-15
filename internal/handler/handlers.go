package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/Hiroki111/go-backend-example/internal/auth"
	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/repository"
	"gorm.io/gorm"
)

const DefaultPageLimit = 20
const MaxPageLimit = 1000

type Handler struct {
	repo *repository.Repository
}

func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
}

func (h *Handler) RegisterAdmin(w http.ResponseWriter, r *http.Request) {
	h.RegisterUser(w, r, domain.AdminRole)
}

func (h *Handler) RegisterCustomer(w http.ResponseWriter, r *http.Request) {
	h.RegisterUser(w, r, domain.CustomerRole)
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request, role domain.UserRole) {
	var data RegisterUserRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	if data.UserName == "" || data.Password == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "username and password required",
		})
		return
	}

	err := h.repo.CreateUser(domain.User{
		UserName: data.UserName,
		Password: data.Password,
		Role:     role,
	})

	if err != nil {
		if err == repository.ErrUserAlreadyExists {
			writeJSON(w, http.StatusConflict, ErrorResponse{
				Error: "user already exists",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "failed to create user",
		})
		return
	}

	writeJSON(w, http.StatusCreated, RegisterUserResponse{
		Status: "user created",
	})
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var data LoginUserRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	if data.UserName == "" || data.Password == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "username and password required",
		})
		return
	}

	user, err := h.repo.GetUserByCredentials(data.UserName, data.Password)
	if err != nil {
		if err == repository.ErrInvalidCredentials || errors.Is(err, gorm.ErrRecordNotFound) {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{
				Error: "invalid username or password",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "failed to find the user",
		})
		return
	}

	token, err := auth.GenerateJWTToken(user.ID, user.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "failed to get token",
		})
		return
	}

	writeJSON(w, http.StatusOK, LoginUserResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	})
}

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	orderBy := r.URL.Query().Get("orderBy")
	sortIn := r.URL.Query().Get("sortIn")
	name := r.URL.Query().Get("name")
	minPrice := r.URL.Query().Get("minPrice")
	maxPrice := r.URL.Query().Get("maxPrice")
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")

	minPriceInt, err := parseOptionalInt64(minPrice, 0)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid minPrice",
		})
		return
	}

	maxPriceInt, err := parseOptionalInt64(maxPrice, math.MaxInt64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid maxPrice",
		})
		return
	}

	pageInt, err := parseOptionalInt(page, 1)
	if err != nil || pageInt <= 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "page must be a positive integer",
		})
		return
	}

	limitInt, err := parseOptionalInt(limit, DefaultPageLimit)
	if err != nil || limitInt <= 0 || limitInt > MaxPageLimit {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "limit must be a positive integer and not exceed " + strconv.Itoa(MaxPageLimit),
		})
		return
	}

	offset := (pageInt - 1) * limitInt
	inputs := repository.GetProductsInput{
		OrderBy:  orderBy,
		SortIn:   sortIn,
		Name:     name,
		MinPrice: minPriceInt,
		MaxPrice: maxPriceInt,
		Offset:   offset,
		Limit:    limitInt,
	}
	products, total, err := h.repo.GetProductsWithTotalCount(inputs)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "failed to get products",
		})
		return
	}
	items := make([]ProductItem, len(products))
	for i, product := range products {
		items[i] = ProductItem{
			ID:         product.ID,
			Name:       product.Name,
			PriceCents: product.PriceCents,
		}
	}

	writeJSON(w, http.StatusOK, GetProductsResponse{
		Items:   items,
		Page:    pageInt,
		Limit:   limitInt,
		Total:   int(total),
		HasNext: pageInt*limitInt < int(total),
	})
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

	if payload.PriceCents < 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "price_cents must be a positive number",
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

	if payload.PriceCents != nil {
		if *payload.PriceCents < 0 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "price_cents must be a positive number",
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
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "internal error",
		})
		return
	}

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

	writeJSON(w, http.StatusOK, DeleteProductResponse{Message: "success"})
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(UserIDKey).(uint)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "internal error",
		})
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	product, err := h.repo.GetProductById(req.ProductID)
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

	err = h.repo.CreateOrder(domain.Order{
		UserID:     userId,
		ProductID:  product.ID,
		PriceCents: product.PriceCents,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to create order"})
		return
	}

	writeJSON(w, http.StatusCreated, CreateOrderResponse{Message: "success"})
}
