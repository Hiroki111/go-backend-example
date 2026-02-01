package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/repository"
	"gorm.io/gorm"
)

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

	err := h.service.CreateUser(domain.User{
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

	user, err := h.service.GetUser(data.UserName, data.Password)
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

	token, err := h.service.GenerateJWTToken(user.ID, user.Role)
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
