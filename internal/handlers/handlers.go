package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/repository"
)

type RegisterUserRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

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
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = h.repo.CreateUser(domain.User{
		UserName: data.UserName,
		Password: data.Password,
	})

	if err != nil {
		http.Error(w, "Failed to create a user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
