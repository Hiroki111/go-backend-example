package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/repository"
	"github.com/Hiroki111/go-backend-example/internal/service"
	"gorm.io/gorm"
)

// RegisterAdmin godoc
// @Summary      Register an Admin
// @Description  Creates a new user with the Admin role.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      RegisterUserRequest  true  "Admin Registration Payload"
// @Success      201   {object}  RegisterUserResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /register-admin [post]
func (h *Handler) RegisterAdmin(w http.ResponseWriter, r *http.Request) {
	h.RegisterUser(w, r, domain.AdminRole)
}

// RegisterCustomer godoc
// @Summary      Register a Customer
// @Description  Creates a new user with the Customer role.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      RegisterUserRequest  true  "Customer Registration Payload"
// @Success      201   {object}  RegisterUserResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /register-customer [post]
func (h *Handler) RegisterCustomer(w http.ResponseWriter, r *http.Request) {
	h.RegisterUser(w, r, domain.CustomerRole)
}

// LoginUser godoc
// @Summary      User Login
// @Description  Authenticates a user and returns a JWT token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        login  body      LoginUserRequest  true  "Login Credentials"
// @Success      200    {object}  LoginUserResponse
// @Failure      400    {object}  ErrorResponse
// @Failure      401    {object}  ErrorResponse
// @Failure      500    {object}  ErrorResponse
// @Router       /login-user [post]
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request, role domain.UserRole) {
	var data RegisterUserRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	if err := h.validate.Struct(data); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: h.formatValidationError(err),
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

	if err := h.validate.Struct(data); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: h.formatValidationError(err),
		})
		return
	}

	user, err := h.service.GetUser(data.UserName, data.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials || errors.Is(err, gorm.ErrRecordNotFound) {
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
