package main

import (
	"net/http"

	"github.com/Hiroki111/go-backend-example/internal/domain"
	"github.com/Hiroki111/go-backend-example/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(handler *handler.Handler) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)

	mux.Get("/ping", handler.Ping)

	// NOTE: "POST /register-admin" should be protected by token in real world
	mux.Post("/register-admin", handler.RegisterAdmin)
	mux.Post("/register-customer", handler.RegisterCustomer)
	mux.Post("/login-user", handler.LoginUser)

	mux.Get("/products", handler.GetProducts)
	mux.Get("/products/{id}", handler.GetProductById)
	mux.Post("/products", handler.RequireRole(domain.AdminRole, handler.CreateProduct))
	mux.Patch("/products/{id}", handler.RequireRole(domain.AdminRole, handler.UpdateProduct))
	mux.Delete("/products/{id}", handler.RequireRole(domain.AdminRole, handler.DeleteProduct))

	return mux
}
