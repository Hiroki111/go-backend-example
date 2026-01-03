package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Hiroki111/go-backend-example/internal/auth"
	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/repository"
	"gorm.io/gorm"
)

type Handler struct {
	repo *repository.Repository
}

func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
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

	token, err := auth.GenerateJWTToken(user.ID)
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

	inputs := repository.GetProductsInput{OrderBy: orderBy, SortIn: sortIn, Name: name}
	products, err := h.repo.GetProducts(inputs)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "failed to get products",
		})
		return
	}
	items := make([]ProductResponse, len(products))
	for i, product := range products {
		items[i] = ProductResponse{
			ID:         product.ID,
			Name:       product.Name,
			PriceCents: product.PriceCents,
		}
	}

	writeJSON(w, http.StatusOK, map[string][]ProductResponse{
		"items": items,
	})
}
